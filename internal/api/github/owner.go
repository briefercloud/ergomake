package github

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/api/auth"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/logger"
)

func (ghr *githubRouter) listReposForOwner(c *gin.Context) {
	owner := c.Param("owner")

	authData, ok := auth.GetAuthData(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	isAuthorized, err := auth.IsAuthorized(c, owner, authData)
	if err != nil {
		logger.Ctx(c).Err(err).
			Str("owner", owner).
			Msg("fail to check if user is authorized to list owner repos")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if !isAuthorized {
		c.JSON(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	environments, err := ghr.db.FindEnvironmentsByOwner(owner, database.FindEnvironmentsOptions{IncludeDeleted: true})
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
	lastDeployedAtByRepo := make(map[string]*time.Time)
	for _, env := range environments {
		lastDeployedAt := lastDeployedAtByRepo[env.Repo]
		if lastDeployedAt != nil {
			if env.CreatedAt.After(*lastDeployedAt) {
				lastDeployedAtByRepo[env.Repo] = &env.CreatedAt
			}
		} else {
			lastDeployedAtByRepo[env.Repo] = &env.CreatedAt
		}

		if env.DeletedAt.Valid {
			continue
		}

		envs, ok := envsCountByRepo[env.Repo]
		if !ok {
			envsCountByRepo[env.Repo] = 1
		} else {
			envsCountByRepo[env.Repo] = envs + 1
		}
	}

	repositories, err := ghr.ghApp.ListOwnerInstalledRepos(c, owner)
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

	repos := []gin.H{}
	for _, repo := range repositories {
		environmentCount := envsCountByRepo[repo.GetName()]
		lastDeployedAt := lastDeployedAtByRepo[repo.GetName()]
		if err != nil {
			logger.Ctx(c).Log().Stack().Err(err).
				Str("owner", owner).
				Str("repo", repo.GetName()).
				Msg("fail to get repo permanent branches")
			c.JSON(
				http.StatusInternalServerError,
				http.StatusText(http.StatusInternalServerError),
			)
			return
		}

		repos = append(repos, gin.H{
			"owner":            owner,
			"name":             repo.GetName(),
			"isInstalled":      true,
			"environmentCount": environmentCount,
			"lastDeployedAt":   lastDeployedAt,
		})
	}

	c.JSON(http.StatusOK, repos)
}
