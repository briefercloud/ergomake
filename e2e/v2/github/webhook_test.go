package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/go-github/v52/github"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"

	"github.com/ergomake/ergomake/e2e/testutils"
	"github.com/ergomake/ergomake/internal/api"
	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/github/ghapp"
	environmentsMocks "github.com/ergomake/ergomake/mocks/environments"
	envvarsMocks "github.com/ergomake/ergomake/mocks/envvars"
	paymentMocks "github.com/ergomake/ergomake/mocks/payment"
	servicelogsMocks "github.com/ergomake/ergomake/mocks/servicelogs"
	usersMocks "github.com/ergomake/ergomake/mocks/users"
)

func findFileUpwards(fileName string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "could not get working directory")
	}

	for {
		filePath := filepath.Join(wd, fileName)
		if _, err := os.Stat(filePath); err == nil {
			return filePath, nil
		}

		// Move up to the parent directory
		parentDir := filepath.Dir(wd)
		if parentDir == wd {
			return "", errors.Errorf("file not found: %s", fileName)
		}

		wd = parentDir
	}
}

var cfg api.Config

func init() {
	configFile, err := findFileUpwards(".env.test.local")
	if err == nil {
		err := godotenv.Load(configFile)
		if err != nil {
			panic(err)
		}
	}

	envconfig.MustProcess("", &cfg)
}

func waitForBuildJobs(clusterClient cluster.Client, jobIDs []string) ([]*batchv1.Job, error) {
	// wait at most 30 seconds for build jobs to be spawned
	timeout := time.After(30 * time.Second)
	for {
		jobs, err := clusterClient.ListJobs(context.Background(), "preview-builds")
		if err != nil {
			return nil, err
		}

		spawnedJobs := 0
		for _, name := range jobIDs {
			for _, job := range jobs {
				if name == job.GetLabels()["preview.ergomake.dev/id"] {
					spawnedJobs += 1
					break
				}
			}
		}

		if len(jobIDs) == spawnedJobs {
			return jobs, nil
		}

		select {
		case <-timeout:
			return nil, errors.Errorf("timed out waiting for build jobs %v to be spawned", jobIDs)
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func waitForAllNamespacesTermination(clusterClient cluster.Client, namespaces []string) error {
	// wait at most 1 minute for all namespaces to be terminated
	timeout := time.After(1 * time.Minute)
	for {
		ns, err := clusterClient.GetPreviewNamespaces(context.Background())

		if err != nil {
			return err
		}

		contained := 0
		for _, n := range ns {
			for _, m := range namespaces {
				if n.GetName() == m {
					contained += 1
					break
				}
			}
		}

		if contained == 0 {
			return nil
		}

		select {
		case <-timeout:
			return errors.Errorf("timed out waiting for namespaces %v to be terminated", namespaces)
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func waitForNewDBEnvironment(db *database.DB, pr int, owner, repo, branch string) (*database.Environment, error) {
	// wait at most 10 seconds for a new environment to be saved to db
	timeout := time.After(10 * time.Second)

	for {
		envs, err := db.FindEnvironmentsByPullRequest(pr, owner, repo, branch, database.FindEnvironmentsByPullRequestOptions{})
		if err != nil {
			return nil, err
		}

		if len(envs) > 0 {
			return &envs[0], nil
		}

		select {
		case <-timeout:
			return nil, errors.New("timed out waiting for an env to be created in the db")
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func waitForServicesInDB(db *database.DB, envID uuid.UUID) ([]database.Service, error) {
	// wait at most 10 seconds for a new environment to be saved to db
	timeout := time.After(10 * time.Second)

	for {
		services, err := db.FindServicesByEnvironment(envID)
		if err != nil {
			return nil, err
		}

		if len(services) > 0 {
			return services, nil
		}

		select {
		case <-timeout:
			return nil, errors.New("timed out waiting for an env to be created in the db")
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func waitForDBEnvironmentDeletion(db *database.DB, pr int, owner, repo, branch string) error {
	// wait at most 10 seconds for all envs to be deleted from db
	timeout := time.After(10 * time.Second)
	for {
		envs, err := db.FindEnvironmentsByPullRequest(pr, owner, repo, branch, database.FindEnvironmentsByPullRequestOptions{})
		if err != nil {
			return err
		}

		if len(envs) == 0 {
			return nil
		}

		select {
		case <-timeout:
			return errors.New("timed out waiting for envs to be deleted from db")
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func TestV2GithubWebhook(t *testing.T) {
	type testCase struct {
		name    string
		headers map[string]string
		payload interface{}
		want    want
		setup   func(
			tc *testCase,
			t *testing.T,
			server *httptest.Server,
			db *database.DB,
			clusterClient cluster.Client,
		)
		assertFn func(
			t *testing.T,
			tc *testCase,
			db *database.DB,
			clusterClient cluster.Client,
		)
	}

	getPr := func(action string) github.PullRequestEvent {
		return github.PullRequestEvent{
			Action: github.String(action),
			Repo: &github.Repository{
				Owner: &github.User{
					Login: github.String("ergomake"),
				},
				Name: github.String("preview-e2e-app"),
			},
			PullRequest: &github.PullRequest{
				Number: github.Int(1),
				Head: &github.PullRequestBranch{
					Ref: github.String("test-pr"),
					SHA: github.String("c1066cbcca4ce1983563e9b4c9dfd7d5257199fc"),
				},
			},
			Sender: &github.User{
				Login: github.String("vieiralucas"),
			},
		}
	}

	tt := []*testCase{
		{
			name:    "sends 401 when missing signature",
			payload: github.PullRequestEvent{},
			want:    want{status: http.StatusUnauthorized},
		},
		{
			name:    "sends 401 when invalid signature",
			headers: map[string]string{"X-Hub-Signature-256": "sha256=invalid_signature"},
			payload: github.PullRequestEvent{},
			want:    want{status: http.StatusUnauthorized},
		},
		{
			name:    "sends 400 when invalid payload",
			headers: map[string]string{"X-Hub-Signature-256": genSignature(t, "not valid")},
			payload: "not valid",
			want:    want{status: http.StatusBadRequest},
		},
		{
			name: "spins up an environment when a pr is opened",
			headers: map[string]string{
				"X-Hub-Signature-256": genSignature(t, getPr("opened")),
				"X-GitHub-Event":      "pull_request",
			},
			payload: getPr("opened"),
			want:    want{status: http.StatusNoContent},
			assertFn: func(
				t *testing.T,
				tc *testCase,
				db *database.DB,
				clusterClient cluster.Client,
			) {
				payload := tc.payload.(github.PullRequestEvent)

				owner := payload.GetRepo().GetOwner().GetLogin()
				repo := payload.GetRepo().GetName()
				branch := payload.GetPullRequest().GetHead().GetRef()
				pr := payload.GetPullRequest().GetNumber()

				env, err := waitForNewDBEnvironment(db, pr, owner, repo, branch)
				require.NoError(t, err)

				services, err := waitForServicesInDB(db, env.ID)
				require.NoError(t, err)

				servicesIDs := []string{}
				for _, s := range services {
					if s.Build != "" {
						servicesIDs = append(servicesIDs, s.ID)
					}
				}

				jobs, err := waitForBuildJobs(clusterClient, servicesIDs)
				require.NoError(t, err)

				// wait for jobs completion for at most 5 minutes
				waitJobsCtx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
				defer cancel()
				res, err := clusterClient.WaitJobs(waitJobsCtx, jobs)
				require.NoError(t, err)
				require.Len(t, res.Failed, 0)

				// wait at most 2 minutes for the ingress to be created
				timeout := time.After(2 * time.Minute)
				var webUrl string
				for {
					url, err := clusterClient.GetIngressUrl(context.Background(), env.ID.String(), "web", "http")
					if err != cluster.ErrIngressNotFound && !assert.NoError(t, err) {
						return
					}

					if url != "" {
						webUrl = url
						break
					}

					select {
					case <-timeout:
						t.Fatal("timed out waiting for the ingress to be created")
						return
					case <-time.After(200 * time.Millisecond):
					}
				}

				browser, err := testutils.NewBrowser()
				require.NoError(t, err)

				// wait at most 1 minute for the app to be avaiable
				timeout = time.After(time.Minute)
				for {
					s, err := browser.LoadURL(webUrl)
					require.NoError(t, err)

					if s == 200 {
						break
					}

					select {
					case <-timeout:
						t.Fatal("timed out waiting for app to be avaiable")
						return
					case <-time.After(400 * time.Millisecond):
					}
				}

				browser.AddTodo("my first todo")
				browser.AddTodo("my second todo")
				browser.WaitSeconds(2)

				browser.Refresh()

				todos, err := browser.GetTodoItems()
				require.NoError(t, err)
				assert.Equal(t, 2, todos)
			},
		},
		{
			name: "terminates the environment when a pr is closed",
			headers: map[string]string{
				"X-Hub-Signature-256": genSignature(t, getPr("closed")),
				"X-GitHub-Event":      "pull_request",
			},
			payload: getPr("closed"),
			want:    want{status: http.StatusNoContent},
			setup: func(
				_ *testCase,
				t *testing.T,
				server *httptest.Server,
				db *database.DB,
				clusterClient cluster.Client,
			) {
				openPr := getPr("opened")

				owner := openPr.GetRepo().GetOwner().GetLogin()
				repo := openPr.GetRepo().GetName()
				branch := openPr.GetPullRequest().GetHead().GetRef()
				pr := openPr.GetPullRequest().GetNumber()

				e := httpexpect.Default(t, server.URL)
				e.POST("/v2/github/webhook").WithHeaders(
					map[string]string{
						"X-Hub-Signature-256": genSignature(t, openPr),
						"X-GitHub-Event":      "pull_request",
					},
				).WithJSON(openPr).Expect().Status(http.StatusNoContent)

				env, err := waitForNewDBEnvironment(db, pr, owner, repo, branch)
				require.NoError(t, err)

				// wait at most 3 minutes for the namespace to be created
				timeout := time.After(2 * time.Minute)
				for {
					namespaces, err := clusterClient.GetPreviewNamespaces(context.Background())
					require.NoError(t, err)

					for _, ns := range namespaces {
						if ns.GetName() == env.ID.String() {
							return
						}
					}

					select {
					case <-timeout:
						t.Fatal("timed out waiting for the ingress to be created")
						return
					case <-time.After(200 * time.Millisecond):
					}
				}
			},
			assertFn: func(
				t *testing.T,
				tc *testCase,
				db *database.DB,
				clusterClient cluster.Client,
			) {
				payload := tc.payload.(github.PullRequestEvent)

				owner := payload.GetRepo().GetOwner().GetLogin()
				repo := payload.GetRepo().GetName()
				branch := payload.GetPullRequest().GetHead().GetRef()
				pr := payload.GetPullRequest().GetNumber()

				envs, err := db.FindEnvironmentsByPullRequest(pr, owner, repo, branch, database.FindEnvironmentsByPullRequestOptions{})
				require.NoError(t, err)
				namespaces := []string{}
				for _, env := range envs {
					namespaces = append(namespaces, env.ID.String())
				}

				err = waitForAllNamespacesTermination(clusterClient, namespaces)
				require.NoError(t, err)

				err = waitForDBEnvironmentDeletion(db, pr, owner, repo, branch)
				assert.NoError(t, err)
			},
		},
		{
			name: "receiving a close event right after a open event does not leave lost previews up forever",
			headers: map[string]string{
				"X-Hub-Signature-256": genSignature(t, getPr("closed")),
				"X-GitHub-Event":      "pull_request",
			},
			payload: getPr("closed"),
			want:    want{status: http.StatusNoContent},
			setup: func(
				_ *testCase,
				t *testing.T,
				server *httptest.Server,
				db *database.DB,
				clusterClient cluster.Client,
			) {
				openPr := getPr("opened")

				owner := openPr.GetRepo().GetOwner().GetLogin()
				repo := openPr.GetRepo().GetName()
				branch := openPr.GetPullRequest().GetHead().GetRef()
				pr := openPr.GetPullRequest().GetNumber()

				e := httpexpect.Default(t, server.URL)
				e.POST("/v2/github/webhook").WithHeaders(
					map[string]string{
						"X-Hub-Signature-256": genSignature(t, openPr),
						"X-GitHub-Event":      "pull_request",
					},
				).WithJSON(openPr).Expect().Status(http.StatusNoContent)

				env, err := waitForNewDBEnvironment(db, pr, owner, repo, branch)
				require.NoError(t, err)

				services, err := waitForServicesInDB(db, env.ID)
				require.NoError(t, err)

				servicesIDs := []string{}
				for _, s := range services {
					if s.Build != "" {
						servicesIDs = append(servicesIDs, s.ID)
					}
				}

				_, err = waitForBuildJobs(clusterClient, servicesIDs)
				require.NoError(t, err)
			},
			assertFn: func(
				t *testing.T,
				tc *testCase,
				db *database.DB,
				clusterClient cluster.Client,
			) {
				payload := tc.payload.(github.PullRequestEvent)

				owner := payload.GetRepo().GetOwner().GetLogin()
				repo := payload.GetRepo().GetName()
				branch := payload.GetPullRequest().GetHead().GetRef()
				pr := payload.GetPullRequest().GetNumber()

				envs, err := db.FindEnvironmentsByPullRequest(pr, owner, repo, branch, database.FindEnvironmentsByPullRequestOptions{})
				require.NoError(t, err)
				require.Greater(t, len(envs), 0)
				namespaces := []string{}
				for _, env := range envs {
					namespaces = append(namespaces, env.ID.String())
				}

				err = waitForDBEnvironmentDeletion(db, pr, owner, repo, branch)
				require.NoError(t, err)

				jobs, err := clusterClient.ListJobs(context.Background(), "preview-builds")
				require.NoError(t, err)
				require.Greater(t, len(jobs), 0)

				// wait for jobs to terminate, don't care if success or failure
				waitCtx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
				defer cancel()
				_, err = clusterClient.WaitJobs(waitCtx, jobs)
				require.NoError(t, err)

				// keep checking that all namespaces are terminated for 30 seconds, we expected no preview namespace
				// to remain or be created at the cluster after this scenario
				timeout := time.After(30 * time.Second)
				for {
					err = waitForAllNamespacesTermination(clusterClient, namespaces)
					require.NoError(t, err)

					select {
					case <-timeout:
						return
					case <-time.After(200 * time.Millisecond):
					}
				}
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			clusterClient, err := cluster.NewK8sClient()
			require.NoError(t, err)

			cfg.GithubWebhookSecret = "secret"
			cfg.Cluster = "minikube"

			db := testutils.CreateRandomDB(t)

			ghApp, err := ghapp.NewGithubClient(cfg.GithubPrivateKey, cfg.GithubAppID)
			require.NoError(t, err)
			apiServer := api.NewServer(db, servicelogsMocks.NewLogStreamer(t), ghApp, clusterClient,
				envvarsMocks.NewEnvVarsProvider(t), environmentsMocks.NewEnvironmentsProvider(t),
				usersMocks.NewService(t), paymentMocks.NewPaymentProvider(t), &cfg)

			server := httptest.NewServer(apiServer)

			if tc.setup != nil {
				tc.setup(tc, t, server, db, clusterClient)
			}

			e := httpexpect.Default(t, server.URL)
			e.POST("/v2/github/webhook").WithHeaders(tc.headers).
				WithJSON(tc.payload).Expect().Status(tc.want.status)

			if tc.assertFn != nil {
				tc.assertFn(t, tc, db, clusterClient)
			}
		})
	}
}
