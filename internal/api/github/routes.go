package github

import (
	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/environments"
	"github.com/ergomake/ergomake/internal/envvars"
	"github.com/ergomake/ergomake/internal/github/ghapp"
	"github.com/ergomake/ergomake/internal/payment"
	"github.com/ergomake/ergomake/internal/privregistry"
)

type githubRouter struct {
	db                   *database.DB
	ghApp                ghapp.GHAppClient
	clusterClient        cluster.Client
	envVarsProvider      envvars.EnvVarsProvider
	privRegistryProvider privregistry.PrivRegistryProvider
	environmentsProvider environments.EnvironmentsProvider
	paymentProvider      payment.PaymentProvider
	webhookSecret        string
	frontendURL          string
}

func NewGithubRouter(
	db *database.DB,
	ghApp ghapp.GHAppClient,
	clusterClient cluster.Client,
	envVarsProvider envvars.EnvVarsProvider,
	privRegistryProvider privregistry.PrivRegistryProvider,
	environmentsProvider environments.EnvironmentsProvider,
	paymentProvider payment.PaymentProvider,
	webhookSecret string,
	frontendURL string,
) *githubRouter {
	return &githubRouter{
		db,
		ghApp,
		clusterClient,
		envVarsProvider,
		privRegistryProvider,
		environmentsProvider,
		paymentProvider,
		webhookSecret,
		frontendURL,
	}
}

func (ghr *githubRouter) AddRoutes(router *gin.RouterGroup) {
	router.POST("/webhook", ghr.webhook)
	router.POST("/marketplace/webhook", ghr.marketplaceWebhook)
	router.GET("/user/organizations", ghr.listUserOrganizations)
	router.GET("/owner/:owner/repos", ghr.listReposForOwner)
}
