package permanentbranches

import (
	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/environments"
	"github.com/ergomake/ergomake/internal/github/ghapp"
	"github.com/ergomake/ergomake/internal/github/ghlauncher"
	"github.com/ergomake/ergomake/internal/permanentbranches"
)

type permanentBranchesRouter struct {
	ghApp                     ghapp.GHAppClient
	ghLaunccher               ghlauncher.GHLauncher
	permanentbranchesProvider permanentbranches.PermanentBranchesProvider
	environmentsProvider      environments.EnvironmentsProvider
}

func NewPermanentBranchesRouter(
	ghApp ghapp.GHAppClient,
	ghLaunccher ghlauncher.GHLauncher,
	permanentbranchesProvider permanentbranches.PermanentBranchesProvider,
	environmentsProvider environments.EnvironmentsProvider,
) *permanentBranchesRouter {
	return &permanentBranchesRouter{ghApp, ghLaunccher, permanentbranchesProvider, environmentsProvider}
}

func (er *permanentBranchesRouter) AddRoutes(router *gin.RouterGroup) {
	router.GET("/owner/:owner/repos/:repo/permanent-branches", er.list)
	router.POST("/owner/:owner/repos/:repo/permanent-branches", er.upsert)
}
