package ghapp

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/die-net/lrucache"
	"github.com/google/go-github/v52/github"
	"github.com/gregjones/httpcache"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ergomake/ergomake/internal/git"
	"github.com/ergomake/ergomake/internal/logger"
	"github.com/ergomake/ergomake/internal/oauthutils"
)

const CheckName string = "Ergomake"

var installationNotFoundError = errors.New("installation not found")
var repoNotFoundError = errors.New("repository not found")

type GHAppClient interface {
	git.RemoteGitClient
	CreateCommitStatus(ctx context.Context, owner, repo, sha, state string, targetURL *string) error
	UpsertComment(ctx context.Context, owner string, repo string, prNumber int, commentID int64, comment string) (*github.IssueComment, error)
	ListOwnerInstalledRepos(ctx context.Context, owner string) ([]*github.Repository, error)
}

type ghAppClient struct {
	*github.Client
	*lrucache.LruCache
}

var cache = lrucache.New(100*1024*1024, 0)

func NewGithubClient(privateKey string, appId int64) (GHAppClient, error) {
	transport := httpcache.NewTransport(cache)
	itr, err := ghinstallation.NewAppsTransport(transport, appId, []byte(privateKey))
	if err != nil {
		return nil, errors.Wrap(err, "fail to create gh app client")
	}

	client := github.NewClient(&http.Client{Transport: itr})

	return &ghAppClient{client, cache}, nil
}

func (gh *ghAppClient) GetCloneToken(ctx context.Context, owner string, repo string) (string, error) {
	return gh.getOwnerInstallationToken(ctx, owner)
}

func (gh *ghAppClient) CloneRepo(ctx context.Context, owner string, repo string, branch string, dir string) error {
	token, err := gh.GetCloneToken(ctx, owner, repo)
	if err != nil {
		return errors.Wrap(err, "fail to get clone token")
	}

	cloneURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", token, owner, repo)
	cmd := exec.Command("git", "clone", "--branch", branch, cloneURL, dir)

	return errors.Wrap(cmd.Run(), "fail to run clone command")
}

func (gh *ghAppClient) GetCloneUrl() string {
	return "https://x-access-token:$(GIT_TOKEN)@github.com/$(OWNER)/$(REPO)"
}

func (gh *ghAppClient) GetCloneParams() []string {
	return []string{
		"--depth", "1",
		"--branch", "$(BRANCH)",
	}
}

func (gh *ghAppClient) GetDefaultBranch(ctx context.Context, owner string, repo string) (string, error) {
	installationClient, err := gh.getOwnerInstallationClient(ctx, owner)
	if err != nil {
		return "", errors.Wrap(err, "failed to create installation client")
	}

	repository, resp, err := installationClient.Repositories.Get(ctx, owner, repo)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return "", repoNotFoundError
		}
		return "", errors.Wrap(err, "failed to get repository")
	}

	return repository.GetDefaultBranch(), nil
}

func (gh *ghAppClient) DoesBranchExist(ctx context.Context, owner string, repo string, branch string) (bool, error) {
	installationClient, err := gh.getOwnerInstallationClient(ctx, owner)
	if err != nil {
		return false, errors.Wrap(err, "failed to create installation client")
	}

	_, resp, err := installationClient.Repositories.GetBranch(ctx, owner, repo, branch, true)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, nil
		}

		return false, errors.Wrap(err, "failed to check branch existence")
	}

	return true, nil
}

func (gh *ghAppClient) CreateCommitStatus(ctx context.Context, owner, repo, sha, state string, targetURL *string) error {
	installationClient, err := gh.getOwnerInstallationClient(ctx, owner)
	if err != nil {
		return errors.Wrap(err, "failed to create installation client")
	}

	repoStatus := &github.RepoStatus{
		State:     github.String(state),
		TargetURL: targetURL,
		Context:   github.String("Ergomake"),
	}

	_, res, err := installationClient.Repositories.CreateStatus(ctx, owner, repo, sha, repoStatus)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusForbidden {
			logger.Ctx(ctx).Warn().AnErr("err", err).Str("state", state).Msg("fail to create commit status, missing permissions")
			return nil
		}

		return errors.Wrap(err, "failed to create repo status")
	}

	return nil
}

func (gh *ghAppClient) PostComment(ctx context.Context, owner string, repo string, prNumber int, comment string) error {
	installationClient, err := gh.getOwnerInstallationClient(ctx, owner)
	if err != nil {
		return errors.Wrap(err, "failed to create installation client")
	}

	commentOptions := github.IssueComment{
		Body: &comment,
	}

	_, _, err = installationClient.Issues.CreateComment(ctx, owner, repo, prNumber, &commentOptions)
	if err != nil {
		return errors.Wrap(err, "failed to post comment")
	}

	return nil
}

func (gh *ghAppClient) UpsertComment(
	ctx context.Context,
	owner string,
	repo string,
	prNumber int,
	commentID int64,
	comment string,
) (*github.IssueComment, error) {
	installationClient, err := gh.getOwnerInstallationClient(ctx, owner)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create installation client")
	}

	ghComment := &github.IssueComment{
		Body: &comment,
	}

	if commentID != 0 {
		ghComment, res, err := installationClient.Issues.EditComment(ctx, owner, repo, commentID, ghComment)
		if err != nil {
			if res != nil && res.StatusCode != http.StatusNotFound {
				return nil, errors.Wrapf(err, "failed to edit comment %d", commentID)
			}
		} else {
			return ghComment, nil
		}
	}

	ghComment, _, err = installationClient.Issues.CreateComment(ctx, owner, repo, prNumber, ghComment)
	return ghComment, errors.Wrap(err, "failed to create comment")
}

func (gh *ghAppClient) ListOwnerInstalledRepos(ctx context.Context, owner string) ([]*github.Repository, error) {
	installationClient, err := gh.getOwnerInstallationClient(ctx, owner)
	if err != nil {
		if errors.Is(err, installationNotFoundError) {
			return make([]*github.Repository, 0), nil
		}

		return nil, errors.Wrapf(err, "fail to get owner %s installation client", owner)
	}

	opt := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}
	var allRepos []*github.Repository
	for {
		repos, resp, err := installationClient.Apps.ListRepos(ctx, opt)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to list owner %s repositories", owner)
		}

		allRepos = append(allRepos, repos.Repositories...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func (gh *ghAppClient) getInstallationForUser(ctx context.Context, username string) (*github.Installation, error) {
	installation, res, err := gh.Apps.FindUserInstallation(ctx, username)

	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil, installationNotFoundError
		}

		return nil, errors.Wrapf(err, "fail to find installation for user %s", username)
	}

	return installation, nil
}

func (gh *ghAppClient) getInstallationForOrg(ctx context.Context, org string) (*github.Installation, error) {
	installation, res, err := gh.Apps.FindOrganizationInstallation(ctx, org)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil, installationNotFoundError
		}

		return nil, errors.Wrapf(err, "fail to find installation for org %s", org)
	}

	return installation, nil
}

func (gh *ghAppClient) getOwnerInstallation(ctx context.Context, owner string) (*github.Installation, error) {
	var installation *github.Installation
	var err error

	installation, err = gh.getInstallationForOrg(ctx, owner)
	if err != nil {
		if err == installationNotFoundError {
			installation, err = gh.getInstallationForUser(ctx, owner)
		}
	}

	return installation, err
}

func (gh *ghAppClient) getOwnerInstallationToken(
	ctx context.Context,
	owner string,
) (string, error) {
	installation, err := gh.getOwnerInstallation(ctx, owner)
	if err != nil {
		return "", errors.Wrapf(err, "fail to get owner %s installation", owner)
	}

	token, _, err := gh.Apps.CreateInstallationToken(ctx, installation.GetID(), &github.InstallationTokenOptions{})
	if err != nil {
		return "", errors.Wrapf(err, "fail to create owner %s installation token", owner)
	}

	return token.GetToken(), nil
}

func (gh *ghAppClient) getOwnerInstallationClient(
	ctx context.Context,
	owner string,
) (*github.Client, error) {
	token, err := gh.getOwnerInstallationToken(ctx, owner)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to get owner %s installation token", owner)
	}

	httpClient := oauthutils.CachedHTTPClient(&oauth2.Token{AccessToken: token}, gh.LruCache)
	installationClient := github.NewClient(httpClient)

	return installationClient, nil
}
