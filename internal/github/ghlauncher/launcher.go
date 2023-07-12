package ghlauncher

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v52/github"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/environments"
	"github.com/ergomake/ergomake/internal/envvars"
	"github.com/ergomake/ergomake/internal/github/ghapp"
	"github.com/ergomake/ergomake/internal/logger"
	"github.com/ergomake/ergomake/internal/privregistry"
	"github.com/ergomake/ergomake/internal/transformer"
)

type LaunchEnvironmentRequest struct {
	Owner       string
	BranchOwner string
	Repo        string
	Branch      string
	SHA         string
	PrNumber    *int
	Author      string
	IsPrivate   bool
}
type GHLauncher interface {
	LaunchEnvironment(ctx context.Context, req LaunchEnvironmentRequest) error
}

type ghLauncher struct {
	db                      *database.DB
	ghApp                   ghapp.GHAppClient
	clusterClient           cluster.Client
	envVarsProvider         envvars.EnvVarsProvider
	privRegistryProvider    privregistry.PrivRegistryProvider
	environmentsProvider    environments.EnvironmentsProvider
	dockerhubPullSecretName string
	frontendURL             string
}

func NewGHLauncher(
	db *database.DB,
	ghApp ghapp.GHAppClient,
	clusterClient cluster.Client,
	envVarsProvider envvars.EnvVarsProvider,
	privRegistryProvider privregistry.PrivRegistryProvider,
	environmentsProvider environments.EnvironmentsProvider,
	dockerhubPullSecretName string,
	frontendURL string,
) *ghLauncher {
	return &ghLauncher{
		db,
		ghApp,
		clusterClient,
		envVarsProvider,
		privRegistryProvider,
		environmentsProvider,
		dockerhubPullSecretName,
		frontendURL,
	}
}

func (gh *ghLauncher) LaunchEnvironment(ctx context.Context, req LaunchEnvironmentRequest) error {
	var previousEnvs []database.Environment
	if req.PrNumber != nil {
		envs, err := gh.db.FindEnvironmentsByPullRequest(
			*req.PrNumber,
			req.Owner,
			req.Repo,
			req.Branch,
			database.FindEnvironmentsOptions{IncludeDeleted: true},
		)
		if err != nil {
			return errors.Wrap(err, "fail to find previous envs of pull request")
		}
		previousEnvs = envs
	} else {
		envs, err := gh.environmentsProvider.ListEnvironmentsByBranch(ctx, req.Owner, req.Repo, req.Branch)
		if err != nil {
			return errors.Wrap(err, "fail to find previous envs of branch")
		}

		for _, env := range envs {
			if env.PullRequest.Valid {
				continue
			}

			previousEnvs = append(previousEnvs, *env)
		}
	}

	previousCommentID := int64(0)
	for _, previousEnv := range previousEnvs {
		if previousEnv.GHCommentID > previousCommentID {
			previousCommentID = previousEnv.GHCommentID
		}
	}

	isLimited, err := gh.environmentsProvider.IsOwnerLimited(ctx, req.Owner)
	if err != nil {
		return errors.Wrap(err, "fail to check if owner is limited")
	}

	uid := uuid.New()

	t := transformer.NewGitCompose(
		gh.clusterClient,
		gh.ghApp,
		gh.db,
		gh.envVarsProvider,
		gh.privRegistryProvider,
		req.Owner,
		req.BranchOwner,
		req.Repo,
		req.Branch,
		req.SHA,
		req.PrNumber,
		req.Author,
		!req.IsPrivate,
		gh.dockerhubPullSecretName,
	)
	defer t.Cleanup()

	prepare, err := t.Prepare(ctx, uid)
	if err != nil {
		return errors.Wrap(err, "fail to prepare repo for transform")
	}

	envFrontendLink := fmt.Sprintf("%s/gh/%s/repos/%s/envs/%s", gh.frontendURL, req.Owner, req.Repo, uid)

	env := prepare.Environment
	if previousCommentID != 0 {
		env.GHCommentID = previousCommentID
		err = gh.db.Save(&env).Error
		if err != nil {
			return errors.Wrap(err, "fail to save previousCommentID to env")
		}
	}

	if isLimited {
		err := gh.ghApp.CreateCommitStatus(ctx, req.Owner, req.Repo, req.SHA, "failure", github.String(envFrontendLink))
		if err != nil {
			logger.Ctx(ctx).Err(err).Str("conclusion", "failure").Msg("fail to create commit status for limited env")
		}

		env := prepare.Environment

		if req.PrNumber != nil {
			comment := createLimitedComment()
			ghComment, err := gh.ghApp.UpsertComment(ctx, req.Owner, req.Repo, *req.PrNumber, previousCommentID, comment)
			if err != nil {
				logger.Ctx(ctx).Err(err).Msg("fail to create gh comment for limited env")
			} else {
				env.GHCommentID = ghComment.GetID()
			}
		}

		env.Status = database.EnvLimited

		err = gh.db.Save(env).Error
		if err != nil {
			return errors.Wrap(err, "fail to save GHCommentID to database for limited env")
		}

		logger.Ctx(ctx).Info().Msg("owner limited")

		return nil
	}

	if prepare.Skip {
		logger.Ctx(ctx).Info().Msg("pr skipped because .ergomake folder was not present")
		return nil
	}

	if prepare.ValidationError != nil {
		FailRun(ctx, gh.ghApp, gh.db, envFrontendLink, prepare.Environment, req.SHA, prepare.ValidationError)
		return nil
	}

	err = gh.ghApp.CreateCommitStatus(ctx, req.Owner, req.Repo, req.SHA, "pending", github.String(envFrontendLink))
	if err != nil {
		return errors.Wrap(err, "fail to create commit status")
	}

	transformResult, err := t.Transform(ctx, uid)

	if err != nil {
		FailRun(ctx, gh.ghApp, gh.db, envFrontendLink, prepare.Environment, req.SHA, nil)
		return errors.Wrap(err, "fail to transform compose into cluster env")
	}

	if transformResult.Failed() {
		FailRun(ctx, gh.ghApp, gh.db, envFrontendLink, prepare.Environment, req.SHA, nil)
		return nil
	}

	// we're done building, is the environment still supposed to be launched?
	// try to find dbEnv in the database, if it is deleted, it is because we
	// are not suppose to launch it anymore
	_, err = gh.db.FindEnvironmentByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = gh.ghApp.CreateCommitStatus(ctx, req.Owner, req.Repo, req.SHA, "failure", nil)
			if err != nil {
				logger.Ctx(ctx).Err(err).Str("conclusion", "failure").Msg("fail to create commit status")
			}
			return nil
		}

		FailRun(ctx, gh.ghApp, gh.db, envFrontendLink, prepare.Environment, req.SHA, nil)
		return errors.Wrap(err, "fail to check if env should still be launched")
	}

	err = cluster.Deploy(ctx, gh.clusterClient, transformResult.ClusterEnv)
	if err != nil {
		FailRun(ctx, gh.ghApp, gh.db, envFrontendLink, prepare.Environment, req.SHA, nil)
		return errors.Wrap(err, "fail to deploy cluster env to cluster")
	}

	if transformResult.IsCompose {
		deploymentsCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
		defer cancel()
		err = gh.clusterClient.WaitDeployments(deploymentsCtx, transformResult.ClusterEnv.Namespace)
		if err != nil {
			FailRun(ctx, gh.ghApp, gh.db, envFrontendLink, prepare.Environment, req.SHA, nil)
			return errors.Wrap(err, "fail to wait for deployments")
		}

		SuccessRun(ctx, gh.ghApp, gh.db, envFrontendLink, transformResult.Environment, prepare.Environment, req.SHA)
	}

	return nil
}

func FailRun(
	ctx context.Context,
	ghApp ghapp.GHAppClient,
	db *database.DB,
	envFrontendLink string,
	env *database.Environment,
	sha string,
	validationError *transformer.ProjectValidationError,
) {
	log := logger.Ctx(ctx)

	if env.PullRequest.Valid {
		comment := createFailureComment(envFrontendLink, validationError)
		ghComment, err := ghApp.UpsertComment(ctx, env.Owner, env.Repo, int(env.PullRequest.Int32), env.GHCommentID, comment)
		if err != nil {
			log.Err(err).Msg("fail to post failure comment")
		} else {
			env.GHCommentID = ghComment.GetID()
			err := db.Save(&env).Error
			if err != nil {
				log.Err(err).Msg("fail to save GHCommentID to database for env")
			}
		}
	}

	err := ghApp.CreateCommitStatus(ctx, env.Owner, env.Repo, sha, "failure", github.String(envFrontendLink))
	if err != nil {
		log.Err(err).Str("conclusion", "failure").Msg("fail to create commit status")
	}
}

func SuccessRun(
	ctx context.Context,
	ghApp ghapp.GHAppClient,
	db *database.DB,
	envFrontendLink string,
	compose *transformer.Environment,
	env *database.Environment,
	sha string,
) {
	log := logger.Ctx(ctx)

	if env.PullRequest.Valid {
		comment := createSuccessComment(compose, envFrontendLink)
		ghComment, err := ghApp.UpsertComment(ctx, env.Owner, env.Repo, int(env.PullRequest.Int32), env.GHCommentID, comment)
		if err != nil {
			log.Err(err).Msg("fail to post success comment")
		} else {
			env.GHCommentID = ghComment.GetID()
			err := db.Save(&env).Error
			if err != nil {
				log.Err(err).Msg("fail to save GHCommentID to database for env")
			}
		}
	}

	err := ghApp.CreateCommitStatus(ctx, env.Owner, env.Repo, sha, "success", github.String(envFrontendLink))
	if err != nil {
		log.Err(err).Str("conclusion", "success").Msg("fail to create commit status")
	}
}
