package variables

import (
	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/envvars"
)

type variablesRouter struct {
	envVarsProvider envvars.EnvVarsProvider
}

func NewVariablesRouter(envVarsProvider envvars.EnvVarsProvider) *variablesRouter {
	return &variablesRouter{envVarsProvider}
}

func (er *variablesRouter) AddRoutes(router *gin.RouterGroup) {
	router.GET("/owner/:owner/repos/:repo/variables", er.list)
	router.POST("/owner/:owner/repos/:repo/variables", er.upsert)
}
