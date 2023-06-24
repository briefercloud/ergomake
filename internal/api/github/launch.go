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

func (r *githubRouter) launchEnvironment(ctx context.Context, event *github.PullRequestEvent) error {
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	branch := event.GetPullRequest().GetHead().GetRef()
	sha := event.GetPullRequest().GetHead().GetSHA()
	prNumber := event.GetPullRequest().GetNumber()

	previousEnvs, err := r.db.FindEnvironmentsByPullRequest(
		prNumber,
		owner,
		repo,
		branch,
		database.FindEnvironmentsByPullRequestOptions{IncludeDeleted: true},
	)
	if err != nil {
		return errors.Wrap(err, "fail to find previous envs")
	}

	previousCommentID := int64(0)
	for _, previousEnv := range previousEnvs {
		if previousEnv.GHCommentID > previousCommentID {
			previousCommentID = previousEnv.GHCommentID
		}
	}

	isLimited, err := r.environmentsProvider.IsOwnerLimited(ctx, owner)
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
		owner,
		repo,
		branch,
		sha,
		prNumber,
		event.GetPullRequest().GetUser().GetLogin(),
	)
	defer t.Cleanup()

	prepare, err := t.Prepare(ctx, uid)
	if err != nil {
		return errors.Wrap(err, "fail to prepare repo for transform")
	}

	envFrontendLink := fmt.Sprintf("%s/gh/%s/repos/%s/envs/%s", r.frontendURL, owner, repo, uid)

	if isLimited {
		err := r.ghApp.CreateCommitStatus(ctx, owner, repo, sha, "failure", github.String(envFrontendLink))
		if err != nil {
			logger.Ctx(ctx).Err(err).Str("conclusion", "failure").Msg("fail to create commit status for limited env")
		}

		env := prepare.Environment

		comment := createLimitedComment()
		ghComment, err := r.ghApp.UpsertComment(ctx, owner, repo, prNumber, previousCommentID, comment)
		if err != nil {
			logger.Ctx(ctx).Err(err).Msg("fail to create gh comment for limited env")
		} else {
			env.GHCommentID = ghComment.GetID()
		}

		env.Status = database.EnvLimited

		err = r.db.Save(env).Error
		return errors.Wrap(err, "fail to save GHCommentID to database for limited env")
	}

	if prepare.Skip {
		logger.Ctx(ctx).Info().Msg("pr skipped because .ergomake folder was not present")
		return nil
	}

	err = r.ghApp.CreateCommitStatus(ctx, owner, repo, sha, "pending", github.String(envFrontendLink))
	if err != nil {
		return errors.Wrap(err, "fail to create check")
	}

	transformResult, err := t.Transform(ctx, uid)

	if err != nil {
		r.failRun(ctx, envFrontendLink, event, prepare.Environment, previousCommentID)
		return errors.Wrap(err, "fail to transform compose into cluster env")
	}

	if transformResult.Failed() {
		r.failRun(ctx, envFrontendLink, event, prepare.Environment, previousCommentID)
		return nil
	}

	// we're done building, is the environment still supposed to be launched?
	// try to find dbEnv in the database, if it is deleted, it is because we
	// are not suppose to launch it anymore
	_, err = r.db.FindEnvironmentByID(uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = r.ghApp.CreateCommitStatus(ctx, owner, repo, sha, "failure", nil)
			if err != nil {
				logger.Ctx(ctx).Err(err).Str("conclusion", "failure").Msg("fail to create commit status")
			}
			return nil
		}

		r.failRun(ctx, envFrontendLink, event, prepare.Environment, previousCommentID)
		return errors.Wrap(err, "fail to check if env should still be launched")
	}

	err = cluster.Deploy(ctx, r.clusterClient, transformResult.ClusterEnv)
	if err != nil {
		r.failRun(ctx, envFrontendLink, event, prepare.Environment, previousCommentID)
		return errors.Wrap(err, "fail to deploy cluster env to cluster")
	}

	deploymentsCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	err = r.clusterClient.WaitDeployments(deploymentsCtx, transformResult.ClusterEnv.Namespace)
	if err != nil {
		r.failRun(ctx, envFrontendLink, event, prepare.Environment, previousCommentID)
		return errors.Wrap(err, "fail to wait for deployments")
	}

	r.successRun(ctx, envFrontendLink, event, transformResult.Compose, prepare.Environment, previousCommentID)
	return nil
}

func (r *githubRouter) failRun(
	ctx context.Context,
	envFrontendLink string,
	event *github.PullRequestEvent,
	env *database.Environment,
	previousCommentID int64,
) {
	log := logger.Ctx(ctx)
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	pr := event.GetPullRequest().GetNumber()
	sha := event.GetPullRequest().GetHead().GetSHA()

	comment := createFailureComment(envFrontendLink)
	ghComment, err := r.ghApp.UpsertComment(ctx, owner, repo, pr, previousCommentID, comment)
	if err != nil {
		log.Err(err).Msg("fail to post failure comment")
	} else {
		env.GHCommentID = ghComment.GetID()
		err := r.db.Save(&env).Error
		if err != nil {
			log.Err(err).Msg("fail to save GHCommentID to database for env")
		}
	}

	err = r.ghApp.CreateCommitStatus(ctx, owner, repo, sha, "failure", github.String(envFrontendLink))
	if err != nil {
		log.Err(err).Str("conclusion", "failure").Msg("fail to create commit status")
	}
}

func (r *githubRouter) successRun(
	ctx context.Context,
	envFrontendLink string,
	event *github.PullRequestEvent,
	compose *transformer.Compose,
	env *database.Environment,
	previousCommentID int64,
) {
	log := logger.Ctx(ctx)
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	pr := event.GetPullRequest().GetNumber()
	sha := event.GetPullRequest().GetHead().GetSHA()

	comment := createSuccessComment(compose)
	ghComment, err := r.ghApp.UpsertComment(ctx, owner, repo, pr, previousCommentID, comment)
	if err != nil {
		log.Err(err).Msg("fail to post success comment")
	} else {
		env.GHCommentID = ghComment.GetID()
		err := r.db.Save(&env).Error
		if err != nil {
			log.Err(err).Msg("fail to save GHCommentID to database for env")
		}
	}

	err = r.ghApp.CreateCommitStatus(ctx, owner, repo, sha, "success", github.String(envFrontendLink))
	if err != nil {
		log.Err(err).Str("conclusion", "success").Msg("fail to create commit status")
	}
}
