package environments

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/ergomake/ergomake/internal/database"
)

var ErrEnvironmentNotFound = errors.New("environment not found")

type TerminateEnvironmentRequest struct {
	Owner    string
	Repo     string
	Branch   string
	PrNumber *int
}

type EnvironmentsProvider interface {
	IsOwnerLimited(ctx context.Context, owner string) (bool, error)
	GetEnvironmentFromHost(ctx context.Context, host string) (*database.Environment, error)
	SaveEnvironment(ctx context.Context, env *database.Environment) error
	ListSuccessEnvironments(ctx context.Context) ([]*database.Environment, error)
	ShouldDeploy(ctx context.Context, owner string, repo string, branch string) (bool, error)
	ListEnvironmentsByBranch(ctx context.Context, owner, repo, branch string) ([]*database.Environment, error)
	DeleteEnvironment(ctx context.Context, id uuid.UUID) error
	TerminateEnvironment(ctx context.Context, req TerminateEnvironmentRequest) error
}
