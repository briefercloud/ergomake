package users

import "context"

type Provider string

const (
	ProviderGithub Provider = "github"
)

type User struct {
	Email    string   `json:"email"`
	Username string   `json:"login"`
	Name     string   `json:"name"`
	Provider Provider `json:"provider"`
}

type Service interface {
	Save(ctx context.Context, user User) error
}
