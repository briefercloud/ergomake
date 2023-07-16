package transformer

import (
	"context"

	"github.com/google/uuid"
	batchv1 "k8s.io/api/batch/v1"

	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
)

type TransformResult struct {
	ClusterEnv  *cluster.ClusterEnv
	Environment *Environment
	FailedJobs  []*batchv1.Job
	IsCompose   bool
}

func (tr *TransformResult) Failed() bool {
	return len(tr.FailedJobs) > 0
}

type PrepareResult struct {
	Environment     *database.Environment
	Skip            bool
	ValidationError *ProjectValidationError
}

type Transformer interface {
	Prepare(ctx context.Context, id uuid.UUID) (*PrepareResult, error)
	Transform(ctx context.Context, namespace string) (*cluster.ClusterEnv, *Environment, error)
}
