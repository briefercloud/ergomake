package github

import (
	"context"

	"github.com/google/go-github/v52/github"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/ergomake/ergomake/internal/database"
)

func (r *githubRouter) terminateEnvironment(ctx context.Context, event *github.PullRequestEvent) error {
	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	branch := event.GetPullRequest().GetHead().GetRef()
	prNumber := event.GetPullRequest().GetNumber()

	envs, err := r.db.FindEnvironmentsByPullRequest(prNumber, owner, repo, branch, database.FindEnvironmentsByPullRequestOptions{})
	if err != nil {
		return errors.Wrap(err, "fail to find environment in DB")
	}

	for _, env := range envs {
		err = r.clusterClient.DeleteNamespace(ctx, env.ID.String())
		if err != nil && !k8sErrors.IsNotFound(err) {
			return errors.Wrap(err, "fail to delete namespace")
		}

		err = r.db.DeleteEnvironmentByPullRequest(prNumber, owner, repo, branch)

		if err != nil {
			return errors.Wrap(err, "fail to delete environment in DB")
		}
	}

	return nil
}
