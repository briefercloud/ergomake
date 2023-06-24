package ghoauth

import (
	"context"
	"net/http"

	"github.com/die-net/lrucache"
	"github.com/google/go-github/v52/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ergomake/ergomake/internal/oauthutils"
)

type GHOAuthClient interface {
	GetUser(ctx context.Context) (*github.User, *github.Response, error)
	ListOrganizations(ctx context.Context) ([]*github.Organization, *github.Response, error)
	ListOwnerRepos(ctx context.Context, owner string) ([]*github.Repository, error)
}

type ghOAuthClient struct {
	*github.Client
}

// 100MB
var cache = lrucache.New(100*1024*1024, 0)

func FromToken(token *oauth2.Token) GHOAuthClient {
	tc := oauthutils.CachedHTTPClient(token, cache)
	client := github.NewClient(tc)

	return &ghOAuthClient{client}
}

func (c *ghOAuthClient) GetUser(ctx context.Context) (*github.User, *github.Response, error) {
	return c.Users.Get(ctx, "")
}

func (c *ghOAuthClient) ListOrganizations(ctx context.Context) ([]*github.Organization, *github.Response, error) {
	var orgs []*github.Organization

	opt := &github.ListOptions{Page: 1, PerPage: 100}
	for {
		pageOrgs, res, err := c.Organizations.List(ctx, "", opt)
		if err != nil {
			return orgs, res, err
		}

		orgs = append(orgs, pageOrgs...)
		if res.NextPage == 0 {
			break
		}
		opt.Page = res.NextPage
	}

	return orgs, nil, nil
}

func (c *ghOAuthClient) isOrg(ctx context.Context, owner string) (bool, error) {
	_, res, err := c.Organizations.Get(ctx, owner)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (c *ghOAuthClient) listUserRepos(ctx context.Context, owner string) ([]*github.Repository, error) {
	authUser, _, err := c.GetUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get authenticated user")
	}

	if authUser.GetLogin() == owner {
		owner = ""
	}

	repositories := []*github.Repository{}
	opt := &github.RepositoryListOptions{ListOptions: github.ListOptions{Page: 1, PerPage: 100}}
	for {
		pageRepos, r, err := c.Repositories.List(ctx, owner, opt)
		if err != nil {
			return nil, err
		}

		repositories = append(repositories, pageRepos...)

		if r.NextPage == 0 {
			break
		}
		opt.Page = r.NextPage
	}

	return repositories, nil
}

func (c *ghOAuthClient) listOrgRepos(ctx context.Context, owner string) ([]*github.Repository, error) {
	repositories := []*github.Repository{}
	opt := &github.RepositoryListByOrgOptions{ListOptions: github.ListOptions{Page: 1, PerPage: 100}}
	for {
		pageRepos, r, err := c.Repositories.ListByOrg(ctx, owner, opt)
		if err != nil {
			return nil, err
		}

		repositories = append(repositories, pageRepos...)

		if r.NextPage == 0 {
			break
		}
		opt.Page = r.NextPage
	}
	return repositories, nil
}

func (c *ghOAuthClient) ListOwnerRepos(ctx context.Context, owner string) ([]*github.Repository, error) {
	ownerIsOrg, err := c.isOrg(ctx, owner)
	if err != nil {
		return nil, errors.Wrapf(err, "failt to check if owner %s is org", owner)
	}

	var repositories []*github.Repository
	if ownerIsOrg {
		repositories, err = c.listOrgRepos(ctx, owner)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to list %s org repos", owner)
		}
	} else {
		repositories, err = c.listUserRepos(ctx, owner)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to list %s user repos", owner)
		}
	}

	return repositories, nil
}
