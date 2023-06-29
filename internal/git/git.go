package git

import (
	"context"
)

type RemoteGitClient interface {
	GetCloneToken(ctx context.Context, owner string, repo string) (string, error)
	CloneRepo(ctx context.Context, owner string, repo string, branch string, dir string, isPublic bool) error
	GetCloneUrl() string
	GetCloneParams() []string
	GetDefaultBranch(ctx context.Context, owner string, repo string) (string, error)
	DoesBranchExist(ctx context.Context, owner string, repo string, branch string) (bool, error)
}
