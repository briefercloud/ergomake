package ghapp

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/die-net/lrucache"
	"github.com/google/go-github/v52/github"
	"github.com/google/uuid"
	"github.com/gregjones/httpcache"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/ergomake/ergomake/internal/git"
	"github.com/ergomake/ergomake/internal/logger"
	"github.com/ergomake/ergomake/internal/oauthutils"
)

const CheckName string = "Ergomake"

var InstallationNotFoundError = errors.New("installation not found")
var repoNotFoundError = errors.New("repository not found")

type GHAppClient interface {
	git.RemoteGitClient
	CreateCommitStatus(ctx context.Context, owner, repo, sha, state string, targetURL *string) error
	UpsertComment(ctx context.Context, owner string, repo string, prNumber int, commentID int64, comment string) (*github.IssueComment, error)
	ListOwnerInstalledRepos(ctx context.Context, owner string) ([]*github.Repository, error)
	IsOwnerInstalled(ctx context.Context, owner string) (bool, error)
	GetInstallation(ctx context.Context, installationID int64) (*github.Installation, error)
	ListInstalledOwners(ctx context.Context) ([]string, error)
	CreatePullRequest(
		ctx context.Context,
		owner, repo, branchPrefix string,
		changes map[string]string,
		title, description string,
	) (*github.PullRequest, error)
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

func (gh *ghAppClient) CloneRepo(ctx context.Context, owner string, repo string, branch string, dir string, isPublic bool) error {
	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
	if !isPublic {
		token, err := gh.GetCloneToken(ctx, owner, repo)
		if err != nil {
			return errors.Wrap(err, "fail to get clone token")
		}

		cloneURL = fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", token, owner, repo)
	}

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

func (gh *ghAppClient) GetDefaultBranch(ctx context.Context, owner string, repo string, branchOwner string) (string, error) {
	installationClient, err := gh.getOwnerInstallationClient(ctx, owner)
	if err != nil {
		return "", errors.Wrap(err, "failed to create installation client")
	}

	repository, resp, err := installationClient.Repositories.Get(ctx, branchOwner, repo)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return "", repoNotFoundError
		}
		return "", errors.Wrap(err, "failed to get repository")
	}

	return repository.GetDefaultBranch(), nil
}

func (gh *ghAppClient) DoesBranchExist(ctx context.Context, owner string, repo string, branch string, branchOwner string) (bool, error) {
	installationClient, err := gh.getOwnerInstallationClient(ctx, owner)
	if err != nil {
		return false, errors.Wrap(err, "failed to create installation client")
	}

	_, resp, err := installationClient.Repositories.GetBranch(ctx, branchOwner, repo, branch, true)
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
		if errors.Is(err, InstallationNotFoundError) {
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
			return nil, InstallationNotFoundError
		}

		return nil, errors.Wrapf(err, "fail to find installation for user %s", username)
	}

	return installation, nil
}

func (gh *ghAppClient) getInstallationForOrg(ctx context.Context, org string) (*github.Installation, error) {
	installation, res, err := gh.Apps.FindOrganizationInstallation(ctx, org)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil, InstallationNotFoundError
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
		if err == InstallationNotFoundError {
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

func (c *ghAppClient) IsOwnerInstalled(ctx context.Context, owner string) (bool, error) {
	_, err := c.getOwnerInstallation(ctx, owner)
	if errors.Is(err, InstallationNotFoundError) {
		return false, nil
	}

	return true, errors.Wrap(err, "fail to get owner installation")
}

func (c *ghAppClient) GetInstallation(ctx context.Context, installationID int64) (*github.Installation, error) {
	installation, res, err := c.Apps.GetInstallation(ctx, installationID)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil, InstallationNotFoundError
		}
	}

	return installation, errors.Wrap(err, "fail to get installation")
}

func (c *ghAppClient) ListInstalledOwners(ctx context.Context) ([]string, error) {
	opt := &github.ListOptions{
		Page:    1,
		PerPage: 100,
	}
	var allInstallations []string
	for {
		installations, resp, err := c.Apps.ListInstallations(ctx, opt)
		if err != nil {
			return nil, errors.Wrap(err, "fail to list installations")
		}

		for _, installation := range installations {
			allInstallations = append(allInstallations, installation.GetAccount().GetLogin())
		}

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return allInstallations, nil
}

func (c *ghAppClient) CreatePullRequest(
	ctx context.Context,
	owner, repo, branchPrefix string,
	changes map[string]string,
	title, description string,
) (*github.PullRequest, error) {
	client, err := c.getOwnerInstallationClient(ctx, owner)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get owner installation client")
	}

	defaultBranch, err := c.GetDefaultBranch(ctx, owner, repo, owner)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to get default branch of repo %s/%s", owner, repo)
	}
	defaultBranchInfo, _, err := client.Repositories.GetBranch(ctx, owner, repo, defaultBranch, true)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to get branch %s of repo %s/%s", defaultBranch, owner, repo)
	}

	branch := branchPrefix
	i := 0
	for {
		if i >= 10 {
			branch = fmt.Sprintf("%s-%s", branchPrefix, uuid.NewString())
		}

		i += 1
		_, res, err := client.Repositories.GetBranch(ctx, owner, repo, branch, true)
		if err == nil {
			branch = fmt.Sprintf("%s-%d", branchPrefix, i)
			continue
		}

		if res != nil && res.StatusCode != http.StatusNotFound {
			return nil, errors.Wrapf(err, "fail to get branch %s of repo %s/%s", branch, owner, repo)
		}

		_, _, err = client.Git.CreateRef(ctx, owner, repo, &github.Reference{
			Ref: github.String(fmt.Sprintf("refs/heads/%s", branch)),
			Object: &github.GitObject{
				Type: github.String("commit"),
				SHA:  defaultBranchInfo.GetCommit().SHA,
			},
		})
		if err != nil {
			return nil, errors.Wrapf(err, "fail to create branch %s at repo %s/%s", branch, owner, repo)
		}
		break
	}

	for path, content := range changes {
		_, _, err := client.Repositories.CreateFile(ctx, owner, repo, path, &github.RepositoryContentFileOptions{
			Message: github.String(fmt.Sprintf("add %s", path)),
			Content: []byte(content),
			Branch:  github.String(branch),
		})

		if err != nil {
			return nil, errors.Wrapf(err, "fail to create file %s into branch %s at repo %s/%s", path, branch, owner, repo)
		}
	}

	pr, _, err := client.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(branch),
		Base:                github.String(defaultBranch),
		Body:                github.String(description),
		MaintainerCanModify: github.Bool(true),
	})

	return pr, errors.Wrapf(err, "fail to create pull request for branch %s at repo %s/%s", branch, owner, repo)
}
