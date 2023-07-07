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
	"github.com/ergomake/ergomake/internal/github/ghapp"
	"github.com/ergomake/ergomake/internal/logger"
	"github.com/ergomake/ergomake/internal/transformer"
)

type LaunchEnvironment struct {
	Owner       string
	BranchOwner string
	Repo        string
	Branch      string
	SHA         string
	PrNumber    *int
	Author      string
	IsPrivate   bool
}

func (r *githubRouter) launchEnvironment(ctx context.Context, event *LaunchEnvironment) error {
	var previousEnvs []database.Environment
	if event.PrNumber != nil {
		envs, err := r.db.FindEnvironmentsByPullRequest(
			*event.PrNumber,
			event.Owner,
			event.Repo,
			event.Branch,
			database.FindEnvironmentsByPullRequestOptions{IncludeDeleted: true},
		)
		if err != nil {
			return errors.Wrap(err, "fail to find previous envs of pull request")
		}
		previousEnvs = envs
	} else {
		envs, err := r.environmentsProvider.ListEnvironmentsByBranch(ctx, event.Owner, event.Repo, event.Branch)
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

	isLimited, err := r.environmentsProvider.IsOwnerLimited(ctx, event.Owner)
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
		event.Owner,
		event.BranchOwner,
		event.Repo,
		event.Branch,
		event.SHA,
		event.PrNumber,
		event.Author,
		!event.IsPrivate,
		r.dockerhubPullSecretName,
	)
	defer t.Cleanup()

	prepare, err := t.Prepare(ctx, uid)
	if err != nil {
		return errors.Wrap(err, "fail to prepare repo for transform")
	}

	envFrontendLink := fmt.Sprintf("%s/gh/%s/repos/%s/envs/%s", r.frontendURL, event.Owner, event.Repo, uid)

	env := prepare.Environment
	if previousCommentID != 0 {
		env.GHCommentID = previousCommentID
		err = r.db.Save(&env).Error
		if err != nil {
			return errors.Wrap(err, "fail to save previousCommentID to env")
		}
	}

	if isLimited {
		err := r.ghApp.CreateCommitStatus(ctx, event.Owner, event.Repo, event.SHA, "failure", github.String(envFrontendLink))
		if err != nil {
			logger.Ctx(ctx).Err(err).Str("conclusion", "failure").Msg("fail to create commit status for limited env")
		}

		env := prepare.Environment

		if event.PrNumber != nil {
			comment := createLimitedComment()
			ghComment, err := r.ghApp.UpsertComment(ctx, event.Owner, event.Repo, *event.PrNumber, previousCommentID, comment)
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
		FailRun(ctx, r.ghApp, r.db, envFrontendLink, prepare.Environment, event.SHA, prepare.ValidationError)
		return nil
	}

	err = r.ghApp.CreateCommitStatus(ctx, event.Owner, event.Repo, event.SHA, "pending", github.String(envFrontendLink))
	if err != nil {
		return errors.Wrap(err, "fail to create commit status")
	}

	transformResult, err := t.Transform(ctx, uid)

	if err != nil {
		FailRun(ctx, r.ghApp, r.db, envFrontendLink, prepare.Environment, event.SHA, nil)
		return errors.Wrap(err, "fail to transform compose into cluster env")
	}

	if transformResult.Failed() {
		FailRun(ctx, r.ghApp, r.db, envFrontendLink, prepare.Environment, event.SHA, nil)
		return nil
	}

	// we're done building, is the environment still supposed to be launched?
	// try to find dbEnv in the database, if it is deleted, it is because we
	// are not suppose to launch it anymore
	_, err = r.db.FindEnvironmentByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = r.ghApp.CreateCommitStatus(ctx, event.Owner, event.Repo, event.SHA, "failure", nil)
			if err != nil {
				logger.Ctx(ctx).Err(err).Str("conclusion", "failure").Msg("fail to create commit status")
			}
			return nil
		}

		FailRun(ctx, r.ghApp, r.db, envFrontendLink, prepare.Environment, event.SHA, nil)
		return errors.Wrap(err, "fail to check if env should still be launched")
	}

	err = cluster.Deploy(ctx, r.clusterClient, transformResult.ClusterEnv)
	if err != nil {
		FailRun(ctx, r.ghApp, r.db, envFrontendLink, prepare.Environment, event.SHA, nil)
		return errors.Wrap(err, "fail to deploy cluster env to cluster")
	}

	if transformResult.IsCompose {
		deploymentsCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
		defer cancel()
		err = r.clusterClient.WaitDeployments(deploymentsCtx, transformResult.ClusterEnv.Namespace)
		if err != nil {
			FailRun(ctx, r.ghApp, r.db, envFrontendLink, prepare.Environment, event.SHA, nil)
			return errors.Wrap(err, "fail to wait for deployments")
		}

		SuccessRun(ctx, r.ghApp, r.db, envFrontendLink, transformResult.Environment, prepare.Environment, event.SHA)
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
		comment := createSuccessComment(compose)
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
