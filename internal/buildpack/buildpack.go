package buildpack

import (
	"context"
	"fmt"

	kpackBuild "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	kpackCore "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"

	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/github/ghapp"
	"github.com/ergomake/ergomake/internal/github/ghlauncher"
	"github.com/ergomake/ergomake/internal/logger"
	"github.com/ergomake/ergomake/internal/transformer"
)

func convertToBuild(obj interface{}) (*kpackBuild.Build, error) {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		err := errors.New("fail to convert newObj of Update event to unstructured.Unstructured")
		return nil, err
	}

	build := &kpackBuild.Build{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		u.Object,
		build,
	)

	return build, errors.Wrapf(
		err,
		"fail to convert %s from Unstructured to Build",
		u.GetName(),
	)
}

func WatchBuilds(
	clusterClient cluster.Client,
	db *database.DB,
	ghApp ghapp.GHAppClient,
	frontendURL string,
) (func(), error) {
	buildCh := make(chan *kpackBuild.Build)
	stopCh := make(chan struct{})

	gvr := schema.GroupVersionResource{
		Group:    "kpack.io",
		Version:  "v1alpha2",
		Resource: "builds",
	}
	starter, err := clusterClient.WatchResource(context.Background(), gvr, cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			oldBuild, err := convertToBuild(oldObj)
			if err != nil {
				logger.Get().Err(err).Msg("fail to convert oldObj to build")
				return
			}

			newBuild, err := convertToBuild(newObj)
			if err != nil {
				logger.Get().Err(err).Msg("fail to convert newObj to build")
				return
			}

			oldCondition := oldBuild.Status.GetCondition(kpackCore.ConditionSucceeded)
			newCondition := newBuild.Status.GetCondition(kpackCore.ConditionSucceeded)

			if newCondition == nil {
				return
			}

			if oldCondition != nil && oldCondition.Status == newCondition.Status {
				return
			}

			if newCondition.IsTrue() || newCondition.IsFalse() {
				buildCh <- newBuild
			}
		},
	})

	go func() {
	outer:
		for {
			select {
			case build := <-buildCh:
				ctx := context.Background()

				condition := build.Status.GetCondition(kpackCore.ConditionSucceeded)
				if condition == nil {
					continue outer
				}

				if condition.IsUnknown() {
					continue outer
				}

				status := "build-success"
				if condition.IsFalse() {
					status = "build-failed"
				}

				labels := build.GetLabels()
				serviceID, ok := labels["preview.ergomake.dev/id"]
				if !ok {
					err := errors.Errorf(
						"preview.ergomake.dev/id label does not exist in kpack build %s",
						build.GetName(),
					)
					logger.Ctx(ctx).Err(err).
						Str("build", build.GetName()).
						Msg("fail to extract service id from kpack build")
					continue outer
				}

				sha := labels["preview.ergomake.dev/sha"]

				var service database.Service
				err := db.First(&service, "id = ?", serviceID).Error
				if err != nil {
					logger.Ctx(ctx).Err(err).Msg("fail to find service in database")
					continue outer
				}

				service.BuildStatus = status
				err = db.Save(&service).Error
				if err != nil {
					logger.Ctx(ctx).Err(err).Msg("fail to update service build status to success in database")
					continue outer
				}

				env, err := db.FindEnvironmentByID(service.EnvironmentID)
				if err != nil {
					if !errors.Is(err, gorm.ErrRecordNotFound) {
						logger.Ctx(ctx).Err(err).Msg("fail to find environment in database")
					}
					continue outer
				}

				envFrontendLink := fmt.Sprintf(
					"%s/gh/%s/repos/%s/envs/%s",
					frontendURL, env.Owner, env.Repo, env.ID.String(),
				)
				success := true
				for _, service := range env.Services {
					if service.BuildStatus == "building" {
						continue outer
					}

					if service.BuildStatus == "image" {
						continue
					}

					success = service.BuildStatus == "build-success"
				}

				if success {
					for _, service := range env.Services {
						err := clusterClient.ScaleDeployment(ctx, env.ID.String(), service.Name, 1)
						if err != nil {
							logger.Ctx(ctx).Err(err).Str("env", env.ID.String()).Str("service", service.Name).
								Msg("fail to scale deployment up when bringing environment up")
							ghlauncher.FailRun(ctx, ghApp, db, envFrontendLink, &env, sha, nil)
							continue outer
						}
					}

					err := db.Model(&env).Update("status", database.EnvSuccess).Error
					if err != nil {
						logger.Ctx(ctx).Err(err).Str("env", env.ID.String()).Msg("fail to update db environment status to success")
						ghlauncher.FailRun(ctx, ghApp, db, envFrontendLink, &env, sha, nil)
						continue outer
					}
					ghlauncher.SuccessRun(ctx, ghApp, db, envFrontendLink, transformer.EnvironmentFromDB(&env), &env, sha)
				} else {
					err := db.Model(&env).Update("status", database.EnvDegraded).Error
					if err != nil {
						logger.Ctx(ctx).Err(err).Str("env", env.ID.String()).Msg("fail to update db environment status to degraded")
					}

					err = clusterClient.DeleteNamespace(ctx, env.ID.String())
					if err != nil {
						logger.Get().Err(err).Str("env", env.ID.String()).Msg("fail to delete namespace after build failure")
						continue outer
					}

					ghlauncher.FailRun(ctx, ghApp, db, envFrontendLink, &env, sha, nil)
				}

				logger.Ctx(ctx).Info().Str("env", env.ID.String()).Bool("success", success).
					Msg("an environment finished building")

			case <-stopCh:
				return
			}
		}
	}()

	starter.Start(stopCh)
	clean := func() {
		close(stopCh)
	}

	return clean, err
}
