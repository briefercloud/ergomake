package github

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
	"github.com/ergomake/ergomake/internal/logger"
	"github.com/ergomake/ergomake/internal/transformer"
)

type LaunchEnvironment struct {
	owner       string
	branchOwner string
	repo        string
	branch      string
	sha         string
	prNumber    *int
	author      string
	isPrivate   bool
}

func (r *githubRouter) launchEnvironment(ctx context.Context, event *LaunchEnvironment) error {
	var previousEnvs []database.Environment
	if event.prNumber != nil {
		envs, err := r.db.FindEnvironmentsByPullRequest(
			*event.prNumber,
			event.owner,
			event.repo,
			event.branch,
			database.FindEnvironmentsByPullRequestOptions{IncludeDeleted: true},
		)
		if err != nil {
			return errors.Wrap(err, "fail to find previous envs")
		}
		previousEnvs = envs
	} else {
		// TODO: search just by branch
		previousEnvs = make([]database.Environment, 0)
	}

	previousCommentID := int64(0)
	for _, previousEnv := range previousEnvs {
		if previousEnv.GHCommentID > previousCommentID {
			previousCommentID = previousEnv.GHCommentID
		}
	}

	isLimited, err := r.environmentsProvider.IsOwnerLimited(ctx, event.owner)
	if err != nil {
		return errors.Wrap(err, "fail to check if owner is limited")
	}

	uid := uuid.New()

	t := transformer.NewGitCompose(
		r.clusterClient,
		r.ghApp,
		r.db,
		r.envVarsProvider,
		r.privRegistryProvider,
		event.owner,
		event.branchOwner,
		event.repo,
		event.branch,
		event.sha,
		event.prNumber,
		event.author,
		!event.isPrivate,
		r.dockerhubPullSecretName,
	)
	defer t.Cleanup()

	prepare, err := t.Prepare(ctx, uid)
	if err != nil {
		return errors.Wrap(err, "fail to prepare repo for transform")
	}

	envFrontendLink := fmt.Sprintf("%s/gh/%s/repos/%s/envs/%s", r.frontendURL, event.owner, event.repo, uid)

	if isLimited {
		err := r.ghApp.CreateCommitStatus(ctx, event.owner, event.repo, event.sha, "failure", github.String(envFrontendLink))
		if err != nil {
			logger.Ctx(ctx).Err(err).Str("conclusion", "failure").Msg("fail to create commit status for limited env")
		}

		env := prepare.Environment

		if event.prNumber != nil {
			comment := createLimitedComment()
			ghComment, err := r.ghApp.UpsertComment(ctx, event.owner, event.repo, *event.prNumber, previousCommentID, comment)
			if err != nil {
				logger.Ctx(ctx).Err(err).Msg("fail to create gh comment for limited env")
			} else {
				env.GHCommentID = ghComment.GetID()
			}
		}

		env.Status = database.EnvLimited

		err = r.db.Save(env).Error
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
		r.failRun(ctx, envFrontendLink, event, prepare.Environment, previousCommentID, prepare.ValidationError)
		return nil
	}

	err = r.ghApp.CreateCommitStatus(ctx, event.owner, event.repo, event.sha, "pending", github.String(envFrontendLink))
	if err != nil {
		return errors.Wrap(err, "fail to create commit status")
	}

	transformResult, err := t.Transform(ctx, uid)

	if err != nil {
		r.failRun(ctx, envFrontendLink, event, prepare.Environment, previousCommentID, nil)
		return errors.Wrap(err, "fail to transform compose into cluster env")
	}

	if transformResult.Failed() {
		r.failRun(ctx, envFrontendLink, event, prepare.Environment, previousCommentID, nil)
		return nil
	}

	// we're done building, is the environment still supposed to be launched?
	// try to find dbEnv in the database, if it is deleted, it is because we
	// are not suppose to launch it anymore
	_, err = r.db.FindEnvironmentByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = r.ghApp.CreateCommitStatus(ctx, event.owner, event.repo, event.sha, "failure", nil)
			if err != nil {
				logger.Ctx(ctx).Err(err).Str("conclusion", "failure").Msg("fail to create commit status")
			}
			return nil
		}

		r.failRun(ctx, envFrontendLink, event, prepare.Environment, previousCommentID, nil)
		return errors.Wrap(err, "fail to check if env should still be launched")
	}

	err = cluster.Deploy(ctx, r.clusterClient, transformResult.ClusterEnv)
	if err != nil {
		r.failRun(ctx, envFrontendLink, event, prepare.Environment, previousCommentID, nil)
		return errors.Wrap(err, "fail to deploy cluster env to cluster")
	}

	deploymentsCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	err = r.clusterClient.WaitDeployments(deploymentsCtx, transformResult.ClusterEnv.Namespace)
	if err != nil {
		r.failRun(ctx, envFrontendLink, event, prepare.Environment, previousCommentID, nil)
		return errors.Wrap(err, "fail to wait for deployments")
	}

	r.successRun(ctx, envFrontendLink, event, transformResult.Compose, prepare.Environment, previousCommentID)
	return nil
}

func (r *githubRouter) failRun(
	ctx context.Context,
	envFrontendLink string,
	event *LaunchEnvironment,
	env *database.Environment,
	previousCommentID int64,
	validationError *transformer.ProjectValidationError,
) {
	log := logger.Ctx(ctx)

	if event.prNumber != nil {
		comment := createFailureComment(envFrontendLink, validationError)
		ghComment, err := r.ghApp.UpsertComment(ctx, event.owner, event.repo, *event.prNumber, previousCommentID, comment)
		if err != nil {
			log.Err(err).Msg("fail to post failure comment")
		} else {
			env.GHCommentID = ghComment.GetID()
			err := r.db.Save(&env).Error
			if err != nil {
				log.Err(err).Msg("fail to save GHCommentID to database for env")
			}
		}
	}

	err := r.ghApp.CreateCommitStatus(ctx, event.owner, event.repo, event.sha, "failure", github.String(envFrontendLink))
	if err != nil {
		log.Err(err).Str("conclusion", "failure").Msg("fail to create commit status")
	}
}

func (r *githubRouter) successRun(
	ctx context.Context,
	envFrontendLink string,
	event *LaunchEnvironment,
	compose *transformer.Compose,
	env *database.Environment,
	previousCommentID int64,
) {
	log := logger.Ctx(ctx)

	if event.prNumber != nil {
		comment := createSuccessComment(compose)
		ghComment, err := r.ghApp.UpsertComment(ctx, event.owner, event.repo, *event.prNumber, previousCommentID, comment)
		if err != nil {
			log.Err(err).Msg("fail to post success comment")
		} else {
			env.GHCommentID = ghComment.GetID()
			err := r.db.Save(&env).Error
			if err != nil {
				log.Err(err).Msg("fail to save GHCommentID to database for env")
			}
		}
	}

	err := r.ghApp.CreateCommitStatus(ctx, event.owner, event.repo, event.sha, "success", github.String(envFrontendLink))
	if err != nil {
		log.Err(err).Str("conclusion", "success").Msg("fail to create commit status")
	}
}
