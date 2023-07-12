package github

import (
	"context"
	"strings"

	"github.com/google/go-github/v52/github"
	"github.com/pkg/errors"

	"github.com/ergomake/ergomake/internal/environments"
	"github.com/ergomake/ergomake/internal/github/ghlauncher"
	"github.com/ergomake/ergomake/internal/logger"
)

func (r *githubRouter) handlePushEvent(githubDelivery string, event *github.PushEvent) {
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo()
	repoName := repo.GetName()
	branch := strings.TrimPrefix(event.GetRef(), "refs/heads/")
	sha := event.GetAfter()
	author := event.GetSender().GetLogin()

	logCtx := logger.With(logger.Get()).
		Str("githubDelivery", githubDelivery).
		Str("owner", owner).
		Str("repo", repoName).
		Str("author", author).
		Str("branch", branch).
		Str("SHA", sha).
		Str("event", "push").
		Logger()
	log := &logCtx
	ctx := log.WithContext(context.Background())

	if _, blocked := ownersBlockList[owner]; blocked {
		log.Warn().Msg("event ignored because owner is in block list")
		return
	}

	log.Info().Msg("got a push event from github")

	shouldDeploy, err := r.environmentsProvider.ShouldDeploy(ctx, owner, repoName, branch)
	if err != nil {
		log.Err(errors.Wrap(err, "fail to check if branch should be deployed")).Msg("fail to handle push event")
		return
	}

	if !shouldDeploy {
		return
	}

	terminateEnv := environments.TerminateEnvironmentRequest{
		Owner:    owner,
		Repo:     repoName,
		Branch:   branch,
		PrNumber: nil,
	}
	err = r.environmentsProvider.TerminateEnvironment(ctx, terminateEnv)
	if err != nil {
		log.Err(err).Msg("fail to terminate environment")
	}

	launchEnv := ghlauncher.LaunchEnvironmentRequest{
		Owner:       owner,
		BranchOwner: owner,
		Repo:        repoName,
		Branch:      branch,
		SHA:         sha,
		PrNumber:    nil,
		Author:      author,
		IsPrivate:   repo.GetPrivate(),
	}

	err = r.launchEnvironment(ctx, launchEnv)
	if err != nil {
		log.Err(err).Msg("fail to launch environment")
	}
}
