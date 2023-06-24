package auth

import (
	"context"

	"github.com/google/go-github/v52/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// TODO: move this to ghoauth.GHOAuthClient
func IsAuthorized(ctx context.Context, owner string, authData *AuthData) (bool, error) {
	tokenSource := oauth2.StaticTokenSource(authData.GithubToken)
	oauth2Client := oauth2.NewClient(ctx, tokenSource)
	client := github.NewClient(oauth2Client)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return false, errors.Wrap(err, "fail to get github authenticated user")
	}

	if *user.Login == owner {
		return true, nil
	}

	isMember, _, err := client.Organizations.IsMember(ctx, owner, *user.Login)
	if err != nil {
		return false, errors.Wrapf(err, "fail to check if user %s is member of org %s", *user.Login, owner)
	}

	return isMember, nil
}
