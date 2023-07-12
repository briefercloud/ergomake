package github

import (
	"context"

	"github.com/ergomake/ergomake/internal/environments"
)

func (r *githubRouter) terminateEnvironment(ctx context.Context, req environments.TerminateEnvironmentRequest) error {
	return r.environmentsProvider.TerminateEnvironment(ctx, req)
}
