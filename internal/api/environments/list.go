package environments

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v52/github"

	"github.com/ergomake/ergomake/internal/api/auth"
	"github.com/ergomake/ergomake/internal/logger"
)

func (er *environmentsRouter) list(c *gin.Context) {
	authData, ok := auth.GetAuthData(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	owner := c.Query("owner")
	repo := c.Query("repo")
	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	isAuthorized, err := auth.IsAuthorized(c, owner, authData)
	if err != nil {
		logger.Ctx(c).Err(err).Msg("fail to create registry")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if !isAuthorized {
		c.JSON(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	ownerEnvs, err := er.db.FindEnvironmentsByOwner(owner)
	if err != nil {
		logger.Ctx(c).Err(err).Msgf("fail to find environments for owner %s", owner)
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	envs := make([]gin.H, 0)
	for _, env := range ownerEnvs {
		if env.Repo != repo {
			continue
		}

		source := "cli"
		if env.PullRequest.Valid {
			source = "pr"
		}

		var branch *string
		if env.Branch.Valid {
			branch = github.String(env.Branch.String)
		}

		services := make([]gin.H, len(env.Services))
		for i, service := range env.Services {
			services[i] = gin.H{
				"id":    service.ID,
				"name":  service.Name,
				"url":   service.Url,
				"build": service.Build,
			}
		}

		envs = append(envs, gin.H{
			"id":        env.ID,
			"branch":    branch,
			"source":    source,
			"status":    env.Status,
			"createdAt": env.CreatedAt,
			"services":  services,
		})
	}

	c.JSON(http.StatusOK, envs)
}
