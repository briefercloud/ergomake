package environments

import (
	"context"
	"errors"

	"github.com/ergomake/ergomake/internal/database"
)

var ErrEnvironmentNotFound = errors.New("environment not found")

type EnvironmentsProvider interface {
	IsOwnerLimited(ctx context.Context, owner string) (bool, error)
	GetEnvironmentFromHost(ctx context.Context, host string) (*database.Environment, error)
	SaveEnvironment(ctx context.Context, env *database.Environment) error
	ListSuccessEnvironments(ctx context.Context) ([]*database.Environment, error)
}
