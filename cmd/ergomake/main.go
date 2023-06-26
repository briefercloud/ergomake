package main

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/ergomake/ergomake/internal/api"
	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/elastic"
	"github.com/ergomake/ergomake/internal/env"
	"github.com/ergomake/ergomake/internal/environments"
	"github.com/ergomake/ergomake/internal/envvars"
	"github.com/ergomake/ergomake/internal/github/ghapp"
	"github.com/ergomake/ergomake/internal/logger"
	"github.com/ergomake/ergomake/internal/payment"
	"github.com/ergomake/ergomake/internal/servicelogs"
	"github.com/ergomake/ergomake/internal/stale"
	"github.com/ergomake/ergomake/internal/users"
)

func main() {
	var cfg api.Config
	err := env.LoadEnv(&cfg)
	if err != nil {
		panic(errors.Wrap(err, "failed to load environment variables"))
	}

	log := logger.Setup()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().AnErr("err", err).Msg("fail to connect to database")
	}

	ghApp, err := ghapp.NewGithubClient(cfg.GithubPrivateKey, cfg.GithubAppID)
	if err != nil {
		log.Fatal().AnErr("err", err).Msg("fail to create GitHub client")
	}

	clusterClient, err := cluster.NewK8sClient()
	if err != nil {
		log.Fatal().AnErr("err", err).Msg("fail to create k8s client")
	}

	es, err := elastic.NewElasticSearch(cfg.ElasticSearchURL, cfg.ElasticSearchUsername, cfg.ElasticSearchPassword)
	if err != nil {
		log.Fatal().AnErr("err", err).Msg("fail to connect to elasticsearch")
	}

	logStreamer := servicelogs.NewESLogStreamer(es, time.Second*5)

	envVarsProvider := envvars.NewDBEnvVarProvider(db, cfg.EnvVarsSecret)

	paymentProvider := payment.NewStripePaymentProvider(db, cfg.StripeSecretKey, cfg.StripeStandardPlanProductID,
		cfg.StripeProfessionalPlanProductID, cfg.Friends, cfg.BestFriends)

	environmentsProvider := environments.NewDBEnvironmentsProvider(db, paymentProvider, cfg.EnvironmentsLimit)

	usersService := users.NewDBUsersService(db)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		api := api.NewServer(db, logStreamer, ghApp, clusterClient, envVarsProvider,
			environmentsProvider, usersService, paymentProvider, &cfg)
		api.Listen(":8080")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		stale := stale.NewServer(
			clusterClient,
			environmentsProvider,
			paymentProvider,
			cfg.FrontendURL,
			time.Hour,
			cfg.IngressNamespace,
			cfg.IngressServiceName,
		)

		var innerWg sync.WaitGroup
		innerWg.Add(2)

		go func() {
			defer innerWg.Done()

			err := stale.Listen(context.Background(), ":9090")
			if err != nil {
				log.Fatal().AnErr("err", err).Msg("fail to run stale server")
			}
		}()

		go func() {
			defer innerWg.Done()

			stale.MonitorStaleServices(context.Background())
		}()

		innerWg.Wait()
	}()

	wg.Wait()
}
