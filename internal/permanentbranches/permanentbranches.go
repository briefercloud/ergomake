package permanentbranches

import "context"

type BatchUpsertResult struct {
	Added   []string
	Removed []string
	Result  []string
}
type PermanentBranchesProvider interface {
	List(ctx context.Context, owner, repo string) ([]string, error)
	IsPermanentBranch(ctx context.Context, owner, repo, branch string) (bool, error)
	BatchUpsert(ctx context.Context, owner, repo string, branches []string) (BatchUpsertResult, error)
}
