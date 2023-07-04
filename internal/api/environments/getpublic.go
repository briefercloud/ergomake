package environments

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ergomake/ergomake/internal/logger"
)

func (er *environmentsRouter) getPublic(c *gin.Context) {
	envID, err := uuid.Parse(c.Param("envID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	env, err := er.db.FindEnvironmentByID(envID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, http.StatusText(http.StatusNotFound))
			return
		}

		logger.Ctx(c).Err(err).Msgf("fail to find environment by id %s", envID)
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	areServicesAlive, err := er.clusterClient.AreServicesAlive(c, env.ID.String())
	if err != nil {
		logger.Ctx(c).Err(err).Msg("fail to check if services are alive")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               env.ID,
		"owner":            env.Owner,
		"repo":             env.Repo,
		"status":           env.Status,
		"areServicesAlive": areServicesAlive,
	})
}
