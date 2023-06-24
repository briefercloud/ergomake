package privregistry

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var ErrRegistryNotFound = errors.New("registry not found")

type RegistryCreds struct {
	ID       uuid.UUID `json:"id"`
	URL      string    `json:"url"`
	Provider string    `json:"provider"`
	Token    string    `json:"-"`
}

type PrivRegistryProvider interface {
	ListCredsByOwner(ctx context.Context, owner string, skipToken bool) ([]RegistryCreds, error)
	FetchCreds(ctx context.Context, owner, image string) (*RegistryCreds, error)
	StoreRegistry(ctx context.Context, owner, url, provider, credentials string) error
	DeleteRegistry(ctx context.Context, id uuid.UUID) error
}
