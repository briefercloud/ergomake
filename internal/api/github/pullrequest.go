package github

import (
	"context"

	"github.com/google/go-github/v52/github"

	"github.com/ergomake/ergomake/internal/logger"
)

func (r *githubRouter) handlePullRequestEvent(githubDelivery string, event *github.PullRequestEvent) {
	action := event.GetAction()

	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo()
	repoName := repo.GetName()
	branchOwner := event.GetPullRequest().GetHead().GetRepo().GetOwner().GetLogin()
	branch := event.GetPullRequest().GetHead().GetRef()
	sha := event.GetPullRequest().GetHead().GetSHA()
	prNumber := event.GetPullRequest().GetNumber()
	author := event.GetSender().GetLogin()

	logCtx := logger.With(logger.Get()).
		Str("githubDelivery", githubDelivery).
		Str("action", action).
		Str("owner", owner).
		Str("repo", repoName).
		Int("prNumber", prNumber).
		Str("author", author).
		Str("branch", branch).
		Str("SHA", sha).
		Logger()
	log := &logCtx
	ctx := log.WithContext(context.Background())

	if _, blocked := ownersBlockList[owner]; blocked {
		log.Warn().Msg("event ignored because owner is in block list")
		return
	}

	terminateEnv := &terminateEnvironment{
		owner:    owner,
		repo:     repoName,
		branch:   branch,
		prNumber: github.Int(prNumber),
	}

	log.Info().Msg("got a pull request event from github")
	switch action {
	case "opened", "reopened", "synchronize":
		err := r.terminateEnvironment(ctx, terminateEnv)
		if err != nil {
			log.Err(err).Msg("fail to terminate environment")
		}

		launchEnv := &LaunchEnvironment{
			Owner:       owner,
			BranchOwner: branchOwner,
			Repo:        repoName,
			Branch:      branch,
			SHA:         sha,
			PrNumber:    &prNumber,
			Author:      author,
			IsPrivate:   repo.GetPrivate(),
		}

		err = r.launchEnvironment(ctx, launchEnv)
		if err != nil {
			log.Err(err).Msg("fail to launch environment")
		}
	case "closed":
		err := r.terminateEnvironment(ctx, terminateEnv)
		if err != nil {
			log.Err(err).Msg("fail to terminate environment")
		}
	}
}
