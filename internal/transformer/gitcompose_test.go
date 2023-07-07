package transformer

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"

	"github.com/ergomake/ergomake/e2e/testutils"
	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/privregistry"
	clusterMock "github.com/ergomake/ergomake/mocks/cluster"
	envvarsMocks "github.com/ergomake/ergomake/mocks/envvars"
	gitMock "github.com/ergomake/ergomake/mocks/git"
	privregistryMock "github.com/ergomake/ergomake/mocks/privregistry"
)

func TestGitCompose_Prepare(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tt := []struct {
		name   string
		errors bool
		setup  func(t *testing.T) *gitCompose
	}{
		{
			name:   "fails when clone repo fails",
			errors: true,
			setup: func(t *testing.T) *gitCompose {
				clusterClient := clusterMock.NewClient(t)
				db := testutils.CreateRandomDB(t)
				gitClient := gitMock.NewRemoteGitClient(t)
				gitClient.EXPECT().CloneRepo(ctx, "owner", "repo", "branch", mock.AnythingOfType("string"), true).
					Return(errors.New("rip"))

				return NewGitCompose(
					clusterClient, gitClient, db,
					envvarsMocks.NewEnvVarsProvider(t),
					privregistryMock.NewPrivRegistryProvider(t),
					"owner", "owner", "repo", "branch", "sha", pointer.Int(1337), "author", true, "hub-secret",
				)
			},
		},
		// TODO: test that it skips when .ergomake not present
		// TODO: test that it creates db at db
		// TODO: test that it loads compose
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gc := tc.setup(t)

			id := uuid.New()

			_, err := gc.Prepare(context.Background(), id)
			if tc.errors {
				assert.Error(t, err)
			}
		})
	}
}

func TestGitCompose_Transform(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name   string
		errors bool
		setup  func(t *testing.T, id uuid.UUID) *gitCompose
	}{
		{
			name:   "creates build jobs when service.Dockerfile is filled but service.Build is empty",
			errors: false,
			setup: func(t *testing.T, id uuid.UUID) *gitCompose {
				clusterClient := clusterMock.NewClient(t)
				gitClient := gitMock.NewRemoteGitClient(t)

				gitClient.EXPECT().GetCloneToken(mock.Anything, "owner", "repo").Return("token", nil)
				gitClient.EXPECT().DoesBranchExist(mock.Anything, "owner", "repo", "branch", "owner").Return(true, nil)
				gitClient.EXPECT().GetCloneUrl().Return("")
				gitClient.EXPECT().GetCloneParams().Return([]string{})
				clusterClient.EXPECT().CreateSecret(
					mock.Anything,
					mock.Anything,
				).Return(nil)

				clusterClient.EXPECT().CreateJob(
					mock.Anything,
					mock.Anything,
				).Return(&batchv1.Job{}, nil)

				clusterClient.EXPECT().WaitJobs(
					mock.Anything,
					mock.Anything,
				).Return(
					&cluster.WaitJobsResult{Failed: []*batchv1.Job{}},
					nil,
				)

				db := testutils.CreateRandomDB(t)

				dbEnv := database.NewEnvironment(id, "owner", "owner", "repo", "branch", pointer.Int(1337), "author", database.EnvPending)
				err := db.Create(&dbEnv).Error
				require.NoError(t, err)

				envVarsProvider := envvarsMocks.NewEnvVarsProvider(t)
				envVarsProvider.EXPECT().ListByRepo(mock.Anything, "owner", "repo").Return(nil, nil)

				privRegistryProvider := privregistryMock.NewPrivRegistryProvider(t)
				privRegistryProvider.EXPECT().FetchCreds(mock.Anything, "owner", "mongo").Return(nil, privregistry.ErrRegistryNotFound)
				privRegistryProvider.EXPECT().FetchCreds(mock.Anything, "owner", "willbuild").Return(nil, privregistry.ErrRegistryNotFound)

				gc := NewGitCompose(
					clusterClient, gitClient, db, envVarsProvider,
					privRegistryProvider,
					"owner", "owner", "repo", "branch", "sha", pointer.Int(1337), "author", false, "hub-secret",
				)
				gc.komposeObject = &kobject.KomposeObject{
					ServiceConfigs: map[string]kobject.ServiceConfig{
						"willbuild": {Build: "", Dockerfile: "Dockerfile"},
						"dontbuild": {Image: "mongo"},
					},
				}
				gc.environment = &Environment{}
				gc.dbEnvironment = dbEnv

				return gc
			},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			id := uuid.New()
			gc := tc.setup(t, id)

			gc.prepared = true
			gc.isCompose = true
			_, err := gc.Transform(context.Background(), id)
			if tc.errors {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitCompose_loadErgopack(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type testCase struct {
		name      string
		setup     func(t *testing.T) *gitCompose
		namespace string
		assertFn  func(t *testing.T, tc *testCase, c *gitCompose, err error)
	}

	repo := uuid.New().String()
	tt := []testCase{
		{
			name: "cleanup function deletes cloned repo dir",
			setup: func(t *testing.T) *gitCompose {
				clusterClient := clusterMock.NewClient(t)
				gitClient := gitMock.NewRemoteGitClient(t)
				gitClient.EXPECT().CloneRepo(ctx, "owner", repo, "branch", mock.AnythingOfType("string"), true).
					Run(func(_ context.Context, _, _, _, dir string, _ bool) {

						composeContent := `
version: '3.5'

services:
  mongo:
    image: mongo
`

						err := ioutil.WriteFile(path.Join(dir, "compose.yaml"), []byte(composeContent), 0644)
						assert.NoError(t, err)
					}).Return(nil)

				return NewGitCompose(
					clusterClient, gitClient, &database.DB{},
					envvarsMocks.NewEnvVarsProvider(t),
					privregistryMock.NewPrivRegistryProvider(t),
					"owner", "owner", repo, "branch", "sha", pointer.Int(1337), "author", true, "hub-secret",
				)
			},
			namespace: "delete-repo",
			assertFn: func(t *testing.T, tc *testCase, c *gitCompose, _ error) {
				c.cleanup()

				tmpDir := os.TempDir()
				prefix := path.Join(fmt.Sprintf("ergomake-%s-%s-%s", c.owner, repo, tc.namespace), repo)
				files, err := ioutil.ReadDir(tmpDir)
				require.NoError(t, err)

				for _, f := range files {
					if f.IsDir() && strings.HasPrefix(f.Name(), prefix) {
						assert.Fail(t, fmt.Sprintf("cloned dir %s%s was not deleted", tmpDir, f.Name()))
						return
					}
				}
			},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gitCompose := tc.setup(t)

			_, err := gitCompose.loadErgopack(context.Background(), tc.namespace)
			require.NoError(t, err)

			if tc.assertFn != nil {
				tc.assertFn(t, &tc, gitCompose, err)
			}
		})
	}
}

func TestGitCompose_makeEnvironmentFromServices(t *testing.T) {
	t.Parallel()

	userlandRegistry = "fake-registry.example.com"

	tt := []struct {
		name       string
		services   map[string]kobject.ServiceConfig
		rawCompose string
		want       map[string]EnvironmentService
	}{
		{
			name: "updates service URL and image",
			services: map[string]kobject.ServiceConfig{
				"service1": {
					Build:                         "path/to/build",
					Image:                         "",
					Name:                          "service1",
					ExposeService:                 "",
					ExposeServiceIngressClassName: "",
					Port: []kobject.Ports{
						{
							ContainerPort: 8080,
							HostPort:      8080,
						},
					},
				},
				"service2": {
					Build:                         "",
					Image:                         "existingImage",
					Name:                          "service2",
					ExposeService:                 "",
					ExposeServiceIngressClassName: "",
				},
			},
			rawCompose: `
version: "3"
services:
  service2:
    image: existingImage
  service1:
    build: path/to/build
    ports:
      - "8080:8080"
`,
			want: map[string]EnvironmentService{
				"service1": {
					Build: "path/to/build",
					Image: "",
					Url:   "service1-owner-repo-1337.env.ergomake.test",
					Index: 1,
				},
				"service2": {
					Build: "",
					Image: "existingImage",
					Url:   "",
					Index: 0,
				},
			},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			clusterClient := clusterMock.NewClient(t)
			gitClient := gitMock.NewRemoteGitClient(t)

			gc := NewGitCompose(
				clusterClient, gitClient, &database.DB{},
				envvarsMocks.NewEnvVarsProvider(t),
				privregistryMock.NewPrivRegistryProvider(t),
				"owner", "owner", "repo", "branch", "sha", pointer.Int(1337), "author", true, "hub-secret",
			)
			env := gc.makeEnvironmentFromKObjectServices(tc.services, tc.rawCompose)

			for k, s := range env.Services {
				assert.NotEmpty(t, s.ID)
				s.ID = ""
				env.Services[k] = s
			}
			assert.Equal(t, tc.want, env.Services)
		})
	}
}

func TestGitCompose_addNodeConstraints(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name string
		obj  *appsv1.Deployment
		want runtime.Object
	}{
		{
			name: "adds node selector and tolerations",
			obj: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "deployment-name",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							NodeSelector: map[string]string{},
							Tolerations:  []corev1.Toleration{},
						},
					},
				},
			},
			want: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "deployment-name",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							NodeSelector: map[string]string{"preview.ergomake.dev/role": "preview"},
							Tolerations: []corev1.Toleration{{
								Key:      "preview.ergomake.dev/domain",
								Operator: "Equal",
								Value:    "previews",
								Effect:   "NoSchedule",
							}},
						},
					},
				},
			},
		},
		{
			name: "works when node selector and tolerations are nil",
			obj: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "deployment-name",
				},
			},
			want: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "deployment-name",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							NodeSelector: map[string]string{"preview.ergomake.dev/role": "preview"},
							Tolerations: []corev1.Toleration{{
								Key:      "preview.ergomake.dev/domain",
								Operator: "Equal",
								Value:    "previews",
								Effect:   "NoSchedule",
							}},
						},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			compose := &gitCompose{}
			compose.addNodeContraints(tc.obj)

			assert.Equal(t, tc.want, tc.obj)
		})
	}
}

func TestGitCompose_getUrl(t *testing.T) {
	t.Parallel()

	c := &gitCompose{
		owner:    "myowner",
		repo:     "myrepo",
		prNumber: pointer.Int(123),
	}

	tt := []struct {
		name     string
		service  kobject.ServiceConfig
		expected string
	}{
		{
			name: "Service with host port",
			service: kobject.ServiceConfig{
				Name: "myservice",
				Port: []kobject.Ports{
					{
						HostPort: 8080,
					},
				},
			},
			expected: fmt.Sprintf("myservice-myowner-myrepo-123.%s", clusterDomain),
		},
		{
			name: "Service without host port",
			service: kobject.ServiceConfig{
				Name: "myservice",
				Port: []kobject.Ports{},
			},
			expected: "",
		},
		{
			name: "Service with uppercase characters",
			service: kobject.ServiceConfig{
				Name: "MyService",
				Port: []kobject.Ports{
					{
						HostPort: 8080,
					},
				},
			},
			expected: fmt.Sprintf("myservice-myowner-myrepo-123.%s", clusterDomain),
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			url := c.getUrl(tc.service)
			assert.Equal(t, tc.expected, url)
		})
	}
}
