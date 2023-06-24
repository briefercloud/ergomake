package environments

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ergomake/ergomake/internal/api/auth"
	"github.com/ergomake/ergomake/internal/logger"
	"github.com/ergomake/ergomake/internal/servicelogs"
)

func (er *environmentsRouter) logs(c *gin.Context, build bool) {
	authData, ok := auth.GetAuthData(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	paramEnvID := c.Param("envID")
	envID, err := uuid.Parse(paramEnvID)
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

		logger.Ctx(c).Err(err).Str("envID", envID.String()).
			Msg("fail to find environment by ID")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	isAuthorized, err := auth.IsAuthorized(c, env.Owner, authData)
	if err != nil {
		logger.Ctx(c).Err(err).Msg("fail to check if caller is authorized")
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if !isAuthorized {
		c.JSON(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	services, err := er.db.FindServicesByEnvironment(env.ID)
	if err != nil {
		logger.Ctx(c).Err(err).Msgf("fail to find services for environment %s", env.ID)
		c.JSON(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	logChan := make(chan []servicelogs.LogEntry)
	errChan := make(chan error)

	if build {
		go er.logStreamer.Stream(c.Request.Context(), services, "preview-builds", logChan, errChan)
	} else {
		go er.logStreamer.Stream(c.Request.Context(), services, env.ID.String(), logChan, errChan)
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	c.Stream(func(_ io.Writer) bool {
		select {
		case logs := <-logChan:
			for _, log := range logs {
				c.SSEvent("log", log)
			}
		case <-time.After(5 * time.Second):
			// send something to prevent timeout
			c.SSEvent("ping", "Ping message")
		case err := <-errChan:
			c.SSEvent("error", "Unexpected error")
			logger.Ctx(c).Err(err).Str("envID", env.ID.String()).Msgf("fail to stream build logs")
			return false
		case <-c.Writer.CloseNotify():
			return false
		}
		return true
	})

}

func (er *environmentsRouter) buildLogs(c *gin.Context) {
	er.logs(c, true)
}

func (er *environmentsRouter) liveLogs(c *gin.Context) {
	er.logs(c, false)
}
