package environments

import (
	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/servicelogs"
)

type environmentsRouter struct {
	db            *database.DB
	logStreamer   servicelogs.LogStreamer
	clusterClient cluster.Client
	jwtSecret     string
}

func NewEnvironmentsRouter(
	db *database.DB,
	logStreamer servicelogs.LogStreamer,
	clusterClient cluster.Client,
	jwtSecret string,
) *environmentsRouter {
	return &environmentsRouter{db, logStreamer, clusterClient, jwtSecret}
}

func (er *environmentsRouter) AddRoutes(router *gin.RouterGroup) {
	router.GET("/", er.list)
	router.GET("/:envID/logs/build", er.buildLogs)
	router.GET("/:envID/logs/live", er.liveLogs)
	router.GET("/:envID/public", er.getPublic)
}
