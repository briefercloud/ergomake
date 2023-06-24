package github

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/api/auth"
	"github.com/ergomake/ergomake/internal/github/ghoauth"
	"github.com/ergomake/ergomake/internal/logger"
)

func (ghr *githubRouter) listReposForOwner(c *gin.Context) {
	owner := c.Param("owner")

	authData, ok := auth.GetAuthData(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	client := ghoauth.FromToken(authData.GithubToken)

	environments, err := ghr.db.FindEnvironmentsByOwner(owner)
	if err != nil {
		logger.Ctx(c).Err(err).
			Str("owner", owner).
			Msg("fail to list org repos")
		c.JSON(
			http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError),
		)
		return
	}

	envsCountByRepo := make(map[string]int)
	for _, env := range environments {
		envs, ok := envsCountByRepo[env.Repo]
		if !ok {
			envsCountByRepo[env.Repo] = 1
		} else {
			envsCountByRepo[env.Repo] = envs + 1
		}
	}

	repositories, err := client.ListOwnerRepos(c, owner)
	if err != nil {
		logger.Ctx(c).Err(err).
			Str("owner", owner).
			Msg("fail to list org repos")
		c.JSON(
			http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError),
		)
		return
	}

	installedReposList, err := ghr.ghApp.ListOwnerInstalledRepos(c, owner)
	if err != nil {
		logger.Ctx(c).Log().Stack().Err(err).
			Str("owner", owner).
			Msgf("fail to get owner installation")
		c.JSON(
			http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError),
		)
		return
	}

	installedRepos := make(map[string]struct{})
	for _, repo := range installedReposList {
		installedRepos[repo.GetName()] = struct{}{}
	}

	repos := []gin.H{}
	for _, repo := range repositories {
		_, isInstalled := installedRepos[repo.GetName()]

		environmentCount, ok := envsCountByRepo[repo.GetName()]
		if !ok {
			environmentCount = 0
		}

		repos = append(repos, gin.H{
			"owner":            owner,
			"name":             repo.GetName(),
			"isInstalled":      isInstalled,
			"environmentCount": environmentCount,
		})
	}

	c.JSON(http.StatusOK, repos)
}
