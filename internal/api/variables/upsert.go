package variables

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/api/auth"
	"github.com/ergomake/ergomake/internal/envvars"
	"github.com/ergomake/ergomake/internal/logger"
)

type upsertVariables = []envvars.EnvVar

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

	var body upsertVariables
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

	toDelete := map[string]struct{}{}
	for _, e := range existingList {
		toDelete[e.Name] = struct{}{}
	}

	for _, v := range body {
		delete(toDelete, v.Name)
		err := vr.envVarsProvider.Upsert(c, owner, repo, v.Name, v.Value)
		if err != nil {
			logger.Ctx(c).Err(err).Msgf("fail to upsert variable %s", v.Name)
			c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
	}

	for name := range toDelete {
		err := vr.envVarsProvider.Delete(c, owner, repo, name)
		if err != nil {
			logger.Ctx(c).Err(err).Msgf("fail to delete variable %s", name)
			c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
	}

	c.JSON(http.StatusOK, body)
}
