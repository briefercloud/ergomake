package github

import (
	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/environments"
	"github.com/ergomake/ergomake/internal/envvars"
	"github.com/ergomake/ergomake/internal/github/ghapp"
	"github.com/ergomake/ergomake/internal/github/ghlauncher"
	"github.com/ergomake/ergomake/internal/payment"
	"github.com/ergomake/ergomake/internal/privregistry"
)

type githubRouter struct {
	ghLauncher              ghlauncher.GHLauncher
	db                      *database.DB
	ghApp                   ghapp.GHAppClient
	clusterClient           cluster.Client
	envVarsProvider         envvars.EnvVarsProvider
	privRegistryProvider    privregistry.PrivRegistryProvider
	environmentsProvider    environments.EnvironmentsProvider
	paymentProvider         payment.PaymentProvider
	webhookSecret           string
	frontendURL             string
	dockerhubPullSecretName string
}

func NewGithubRouter(
	ghLauncher ghlauncher.GHLauncher,
	db *database.DB,
	ghApp ghapp.GHAppClient,
	clusterClient cluster.Client,
	envVarsProvider envvars.EnvVarsProvider,
	privRegistryProvider privregistry.PrivRegistryProvider,
	environmentsProvider environments.EnvironmentsProvider,
	paymentProvider payment.PaymentProvider,
	webhookSecret string,
	frontendURL string,
	dockerhubPullSecretName string,
) *githubRouter {
	return &githubRouter{
		ghLauncher,
		db,
		ghApp,
		clusterClient,
		envVarsProvider,
		privRegistryProvider,
		environmentsProvider,
		paymentProvider,
		webhookSecret,
		frontendURL,
		dockerhubPullSecretName,
	}
}

func (ghr *githubRouter) AddRoutes(router *gin.RouterGroup) {
	router.POST("/webhook", ghr.webhook)
	router.POST("/marketplace/webhook", ghr.marketplaceWebhook)
	router.GET("/user/organizations", ghr.listUserOrganizations)
	router.GET("/owner/:owner/repos", ghr.listReposForOwner)
	router.POST("/owner/:owner/repos/:repo/configure", ghr.configureRepo)
}
