package permanentbranches

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/api/auth"
	"github.com/ergomake/ergomake/internal/environments"
	"github.com/ergomake/ergomake/internal/github/ghapp"
	"github.com/ergomake/ergomake/internal/github/ghlauncher"
	"github.com/ergomake/ergomake/internal/github/ghoauth"
	"github.com/ergomake/ergomake/internal/logger"
)

type upsertPermanentBranches struct {
	Branches []string `json:"branches"`
}

func (pbr *permanentBranchesRouter) upsert(c *gin.Context) {
	authData, ok := auth.GetAuthData(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	owner := c.Param("owner")
	repoStr := c.Param("repo")
	if owner == "" || repoStr == "" {
		c.JSON(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	isAuthorized, err := auth.IsAuthorized(c, owner, authData)
	if err != nil {
		logger.Ctx(c).Err(err).Msg("fail to check for authorization")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if !isAuthorized {
		c.JSON(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	client := ghoauth.FromToken(authData.GithubToken)
	user, _, err := client.GetUser(c)
	if err != nil {
		logger.Ctx(c).Err(err).
			Msg("fail to get authenticated user")
		c.JSON(
			http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError),
		)
		return
	}

	var body upsertPermanentBranches
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "malformed-payload"})
		return
	}

	branches, err := pbr.permanentbranchesProvider.BatchUpsert(c, owner, repoStr, body.Branches)
	if err != nil {
		logger.Ctx(c).Err(err).Msgf("fail to upsert permanent branches for repo %s/%s", owner, repoStr)
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	go func() {
		logCtx := logger.With(logger.Get()).
			Str("owner", owner).
			Str("repo", repoStr).
			Logger()
		log := &logCtx
		ctx := log.WithContext(context.Background())

		var wg sync.WaitGroup
		for _, removed := range branches.Removed {
			wg.Add(1)
			go func(branch string) {
				defer wg.Done()

				req := environments.TerminateEnvironmentRequest{
					Owner:    owner,
					Repo:     repoStr,
					Branch:   branch,
					PrNumber: nil,
				}
				err := pbr.environmentsProvider.TerminateEnvironment(ctx, req)
				if err != nil {
					log.Err(err).Str("branch", branch).
						Msg("fail to terminate environment after branch was removed from permanent branches")
				}
			}(removed)
		}

		for _, added := range branches.Added {
			wg.Add(1)
			go func(branchStr string) {
				defer wg.Done()
				sha, err := pbr.ghApp.GetBranchSHA(ctx, owner, repoStr, branchStr)
				if err != nil {
					if errors.Is(err, ghapp.BranchNotFoundError) {
						return
					}

					logger.Ctx(ctx).Err(err).
						Str("owner", owner).
						Str("repo", repoStr).
						Str("branch", branchStr).
						Msg("fail to get branch sha")
					return
				}

				isPrivate, err := pbr.ghApp.IsRepoPrivate(ctx, owner, repoStr)
				if err != nil {
					if errors.Is(err, ghapp.RepoNotFoundError) {
						return
					}

					logger.Ctx(ctx).Err(err).
						Str("owner", owner).
						Str("repo", repoStr).
						Str("branch", branchStr).
						Msg("fail to check if repo is private")
					return
				}

				req := ghlauncher.LaunchEnvironmentRequest{
					Owner:       owner,
					BranchOwner: owner,
					Repo:        repoStr,
					Branch:      branchStr,
					SHA:         sha,
					Author:      user.GetLogin(),
					IsPrivate:   isPrivate,
				}
				err = pbr.ghLaunccher.LaunchEnvironment(ctx, req)
				if err != nil {
					log.Err(err).Str("branch", branchStr).
						Msg("fial to launch environment after branch was added from permanent branches")
				}
			}(added)
		}
		wg.Wait()
	}()

	output := make([]gin.H, 0)
	for _, b := range branches.Result {
		output = append(output, gin.H{"name": b})
	}
	c.JSON(http.StatusOK, output)
}
