package transformer

import (
	"context"

	"github.com/ergomake/ergomake/internal/cluster"
)

type Transformer interface {
	Transform(ctx context.Context, namespace string) (*cluster.ClusterEnv, *Environment, error)
}
