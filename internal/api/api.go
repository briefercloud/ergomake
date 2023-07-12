package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ergomake/ergomake/internal/api/auth"
	environmentsApi "github.com/ergomake/ergomake/internal/api/environments"
	"github.com/ergomake/ergomake/internal/api/github"
	permanentbranchesApi "github.com/ergomake/ergomake/internal/api/permanentbranches"
	"github.com/ergomake/ergomake/internal/api/registries"
	"github.com/ergomake/ergomake/internal/api/stripe"
	"github.com/ergomake/ergomake/internal/api/variables"
	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/environments"
	"github.com/ergomake/ergomake/internal/envvars"
	"github.com/ergomake/ergomake/internal/github/ghapp"
	"github.com/ergomake/ergomake/internal/github/ghlauncher"
	"github.com/ergomake/ergomake/internal/logger"
	"github.com/ergomake/ergomake/internal/payment"
	"github.com/ergomake/ergomake/internal/permanentbranches"
	"github.com/ergomake/ergomake/internal/privregistry"
	"github.com/ergomake/ergomake/internal/servicelogs"
	"github.com/ergomake/ergomake/internal/users"
)

type Config struct {
	GithubWebhookSecret             string   `split_words:"true"`
	GithubPrivateKey                string   `split_words:"true"`
	GithubAppID                     int64    `split_words:"true"`
	GithubClientID                  string   `split_words:"true"`
	GithubClientSecret              string   `split_words:"true"`
	DatabaseURL                     string   `split_words:"true"`
	Cluster                         string   `split_words:"true"`
	AuthRedirectURL                 string   `split_words:"true"`
	JWTSecret                       string   `split_words:"true"`
	AllowOrigin                     string   `split_words:"true"`
	ElasticSearchURL                string   `split_words:"true"`
	ElasticSearchUsername           string   `split_words:"true"`
	ElasticSearchPassword           string   `split_words:"true"`
	FrontendURL                     string   `split_words:"true"`
	EnvVarsSecret                   string   `split_words:"true"`
	PrivRegistriesSecret            string   `split_words:"true"`
	EnvironmentsLimit               int      `split_words:"true"`
	StripeSecretKey                 string   `split_words:"true"`
	StripeWebhookSecret             string   `split_words:"true"`
	StripeStandardPlanProductID     string   `split_words:"true"`
	StripeProfessionalPlanProductID string   `split_words:"true"`
	IngressNamespace                string   `split_words:"true"`
	IngressServiceName              string   `split_words:"true"`
	Friends                         []string `split_words:"true"`
	BestFriends                     []string `split_words:"true"`
	DockerhubPullSecretName         string   `split_words:"true"`
}

type server struct {
	*gin.Engine
}

func corsMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func NewServer(
	ghLauncher ghlauncher.GHLauncher,
	privRegistryProvider privregistry.PrivRegistryProvider,
	db *database.DB,
	logStreamer servicelogs.LogStreamer,
	ghApp ghapp.GHAppClient,
	clusterClient cluster.Client,
	envVarsProvider envvars.EnvVarsProvider,
	environmentsProvider environments.EnvironmentsProvider,
	usersService users.Service,
	paymentProvider payment.PaymentProvider,
	permanentBranchesProvider permanentbranches.PermanentBranchesProvider,
	cfg *Config,
) *server {
	router := gin.New()

	router.Use(corsMiddleware(cfg.AllowOrigin))

	router.Use(gin.Recovery())
	logger.Middleware(router)
	router.Use(auth.ExtractAuthDataMiddleware(cfg.JWTSecret))

	v2 := router.Group("/v2")
	v2.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	ghRouter := github.NewGithubRouter(
		ghLauncher,
		db,
		ghApp,
		clusterClient,
		envVarsProvider,
		privRegistryProvider,
		environmentsProvider,
		paymentProvider,
		cfg.GithubWebhookSecret,
		cfg.FrontendURL,
		cfg.DockerhubPullSecretName,
	)
	ghRouter.AddRoutes(v2.Group("/github"))

	stripeProvider := payment.NewStripePaymentProvider(
		db, cfg.StripeSecretKey, cfg.StripeStandardPlanProductID, cfg.StripeProfessionalPlanProductID,
		cfg.Friends, cfg.BestFriends)
	stripeRouter := stripe.NewStripeRouter(stripeProvider, cfg.StripeWebhookSecret)
	stripeRouter.AddRoutes(v2.Group("/stripe"))

	authRouter := auth.NewAuthRouter(
		cfg.GithubClientID,
		cfg.GithubClientSecret,
		cfg.AuthRedirectURL,
		cfg.JWTSecret,
		cfg.Cluster != "" && cfg.Cluster != "minikube",
		usersService,
		cfg.FrontendURL,
		ghApp,
	)
	authRouter.AddRoutes(v2.Group("/auth"))

	registriesRouter := registries.NewRegistriesRouter(privRegistryProvider)
	registriesRouter.AddRoutes(v2)

	environmentsRouter := environmentsApi.NewEnvironmentsRouter(db, logStreamer, clusterClient, cfg.JWTSecret)
	environmentsRouter.AddRoutes(v2.Group("/environments"))

	variablesRouter := variables.NewVariablesRouter(envVarsProvider)
	variablesRouter.AddRoutes(v2)

	permanentbranchesRouter := permanentbranchesApi.NewPermanentBranchesRouter(
		ghApp,
		ghLauncher,
		permanentBranchesProvider,
		environmentsProvider,
	)
	permanentbranchesRouter.AddRoutes(v2)

	return &server{router}
}

func (s *server) Listen(addr string) {
	s.Run(addr)
}
