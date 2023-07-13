package variables

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/api/auth"
	"github.com/ergomake/ergomake/internal/envvars"
	"github.com/ergomake/ergomake/internal/logger"
)

func (vr *variablesRouter) upsert(c *gin.Context) {
	authData, ok := auth.GetAuthData(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")
	if owner == "" || repo == "" {
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

	var body []envvars.EnvVar
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "malformed-payload"})
		return
	}

	existingList, err := vr.envVarsProvider.ListByRepo(c, owner, repo)
	if err != nil {
		logger.Ctx(c).Err(err).Msgf("fail to list variables for repo %s/%s", owner, repo)
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	toKeep := make(map[string]bool)
	for _, v := range body {
		err := vr.envVarsProvider.Upsert(c, owner, repo, v.Name, v.Value, v.Branch)
		if err != nil {
			logger.Ctx(c).Err(err).Msgf("fail to upsert variable %s", v.Name)
			c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}

		branch := ""
		if v.Branch != nil {
			branch = *v.Branch
		}

		toKeep[fmt.Sprintf("%s/%s", v.Name, branch)] = true
	}

	for _, v := range existingList {
		branch := ""
		if v.Branch != nil {
			branch = *v.Branch
		}

		key := fmt.Sprintf("%s/%s", v.Name, branch)
		if !toKeep[key] {
			err := vr.envVarsProvider.Delete(c, owner, repo, v.Name, v.Branch)
			if err != nil {
				logger.Ctx(c).Err(err).Msgf("fail to delete variable %s", v.Name)
				c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				return
			}
		}
	}

	c.JSON(http.StatusOK, body)
}
