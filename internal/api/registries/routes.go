package registries

import (
	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/privregistry"
)

type registriesRouter struct {
	privRegistryProvider privregistry.PrivRegistryProvider
}

func NewRegistriesRouter(privRegistryProvider privregistry.PrivRegistryProvider) *registriesRouter {
	return &registriesRouter{privRegistryProvider}
}

func (rr *registriesRouter) AddRoutes(router *gin.RouterGroup) {
	router.POST("/owner/:owner/registries", rr.create)
	router.GET("/owner/:owner/registries", rr.list)
	router.DELETE("/owner/:owner/registries/:registryID", rr.del)
}
