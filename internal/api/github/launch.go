package github

import (
	"context"

	"github.com/ergomake/ergomake/internal/github/ghlauncher"
)

func (r *githubRouter) launchEnvironment(ctx context.Context, event ghlauncher.LaunchEnvironmentRequest) error {
	return r.ghLauncher.LaunchEnvironment(ctx, event)
}
