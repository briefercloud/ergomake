package github

import (
	"context"

	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/ergomake/ergomake/internal/database"
)

type terminateEnvironment struct {
	owner    string
	repo     string
	branch   string
	prNumber *int
}

func (r *githubRouter) terminateEnvironment(ctx context.Context, event *terminateEnvironment) error {
	branchEnvs, err := r.environmentsProvider.ListEnvironmentsByBranch(ctx, event.owner, event.repo, event.branch)
	if err != nil {
		return errors.Wrap(err, "fail to list environments by branch")
	}

	envs := make([]*database.Environment, 0)
	for _, env := range branchEnvs {
		if event.prNumber != nil {
			if env.PullRequest.Valid && env.PullRequest.Int32 == int32(*event.prNumber) {
				envs = append(envs, env)
			}
		} else {
			envs = append(envs, env)
		}
	}

	for _, env := range envs {
		err = r.clusterClient.DeleteNamespace(ctx, env.ID.String())
		if err != nil && !k8sErrors.IsNotFound(err) {
			return errors.Wrap(err, "fail to delete namespace")
		}

		err = r.environmentsProvider.DeleteEnvironment(ctx, env.ID)
		if err != nil {
			return errors.Wrap(err, "fail to delete environment in DB")
		}
	}

	return nil
}
