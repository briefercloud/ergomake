package builder

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/envvars"
	"github.com/ergomake/ergomake/internal/git"
	"github.com/ergomake/ergomake/internal/transformer"

	kpackBuild "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	kpackCore "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
)

var insecureRegistry string

func init() {
	cluster := os.Getenv("CLUSTER")
	if cluster != "eks" {
		insecureRegistry = "host.minikube.internal:5001"
		return
	}
}

type BuildImagesResult struct {
	FailedJobs []*batchv1.Job
}

func (bir *BuildImagesResult) Failed() bool {
	return len(bir.FailedJobs) > 0
}

type GitOptions struct {
	Owner       string
	BranchOwner string
	Repo        string
	Branch      string
	SHA         string
	PrNumber    *int
	Author      string
	IsPublic    bool
}

type builder struct {
	clusterClient      cluster.Client
	gitClient          git.RemoteGitClient
	envVarsProvider    envvars.EnvVarsProvider
	db                 *database.DB
	s3Bucket           string
	awsAccessKey       string
	awsSecretAccessKey string

	projectPath             string
	configFilePath          string
	dbEnvironment           *database.Environment
	isCompose               bool
	komposeObject           *kobject.KomposeObject
	environment             *transformer.Environment
	dockerhubPullSecretName string
}

func NewBuilder(
	clusterClient cluster.Client,
	gitClient git.RemoteGitClient,
	envVarsProvider envvars.EnvVarsProvider,
	db *database.DB,
	s3Bucket string,
	awsAccessKey string,
	awsSecretAccessKey string,
	projectPath string,
	configFilePath string,
	dbEnvironment *database.Environment,
	isCompose bool,
	komposeObject *kobject.KomposeObject,
	environment *transformer.Environment,
	dockerhubPullSecretName string,
) *builder {
	return &builder{
		clusterClient,
		gitClient,
		envVarsProvider,
		db,
		s3Bucket,
		awsAccessKey,
		awsSecretAccessKey,
		projectPath,
		configFilePath,
		dbEnvironment,
		isCompose,
		komposeObject,
		environment,
		dockerhubPullSecretName,
	}
}

func ComputeRepoAndBuildPath(projectPath, configFilePath, buildPath, defaultRepo string) (string, string) {
	projectPath = filepath.Clean(projectPath)
	fullBuildPath := filepath.Clean(path.Join(path.Dir(configFilePath), buildPath))

	projectPathParts := strings.Split(projectPath, string(filepath.Separator))
	buildPathParts := strings.Split(fullBuildPath, string(filepath.Separator))

	minLen := len(projectPathParts)
	if len(buildPathParts) < minLen {
		minLen = len(buildPathParts)
	}

	for i := 0; i < minLen; i++ {
		if projectPathParts[i] != buildPathParts[i] {
			rest := strings.Join(buildPathParts[i+1:], string(filepath.Separator))
			if rest == "" {
				rest = "."
			}

			return buildPathParts[i], rest
		}

	}

	return defaultRepo, buildPath
}

func (b *builder) BuildImagesFromGit(
	ctx context.Context,
	namespace string,
	gitOptions GitOptions,
) (*BuildImagesResult, error) {
	b.dbEnvironment.Status = database.EnvBuilding
	err := b.db.Save(b.dbEnvironment).Error
	if err != nil {
		return nil, errors.Wrap(err, "fail to set env status to building")
	}

	if b.isCompose {
		return b.buildImagesWithKanikoFromGit(ctx, namespace, gitOptions)
	} else {
		return b.buildImagesWithBuildpacksFromGit(ctx, namespace, gitOptions)
	}
}

func (b *builder) BuildImagesFromTar(
	ctx context.Context,
	namespace string,
	tarpath string,
) (*BuildImagesResult, error) {
	b.dbEnvironment.Status = database.EnvBuilding
	err := b.db.Save(b.dbEnvironment).Error
	if err != nil {
		return nil, errors.Wrap(err, "fail to set env status to building")
	}

	if !b.isCompose {
		return nil, errors.New("buildpacks are not supported when building from tar")
	}

	return b.buildImagesWithKanikoFromTar(ctx, namespace, tarpath)
}

func (b *builder) buildImagesWithKanikoFromTar(
	ctx context.Context,
	namespace string,
	tarpath string,
) (*BuildImagesResult, error) {
	jobs := []*batchv1.Job{}
	for k, service := range b.komposeObject.ServiceConfigs {
		if service.Build == "" && service.Dockerfile == "" {
			continue
		}

		_, buildPath := ComputeRepoAndBuildPath(b.projectPath, b.configFilePath, service.Build, "")

		spec := b.makeJobSpec(
			b.environment.Services[k].ID,
			k,
			service,
			buildPath,
			[]envvars.EnvVar{},
			GitOptions{},
			fmt.Sprintf("s3://%s/%s/archive.tar.gz", b.s3Bucket, namespace),
		)
		job, err := b.clusterClient.CreateJob(ctx, spec)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to create build job for service %s", k)
		}
		jobs = append(jobs, job)
	}

	jobCtx, cancelFn := context.WithTimeout(ctx, time.Hour)
	defer cancelFn()
	result, err := b.clusterClient.WaitJobs(jobCtx, jobs)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to wait for build jobs to complete")
	}

	return &BuildImagesResult{result.Failed}, nil
}

func (b *builder) buildImagesWithBuildpacksFromGit(ctx context.Context, namespace string, gitOptions GitOptions) (*BuildImagesResult, error) {
	cloneTokenSecrets := make(map[string]*string)
	builds := make([]*kpackBuild.Build, 0)

	for serviceName, service := range b.environment.Services {
		if service.Build == "" {
			continue
		}

		repo, buildPath := ComputeRepoAndBuildPath(b.projectPath, b.configFilePath, service.Build, gitOptions.Repo)
		if repo == gitOptions.Repo {
			buildPath, _ = filepath.Rel("/", path.Clean(path.Join("/", ".ergomake", buildPath)))
		} else {
			buildPath, _ = filepath.Rel(path.Join("/", repo), path.Clean(path.Join("/", gitOptions.Repo, ".ergomake", buildPath)))
		}

		cloneTokenSecretName, ok := cloneTokenSecrets[repo]
		if !ok && !gitOptions.IsPublic {
			cloneToken, err := b.gitClient.GetCloneToken(ctx, gitOptions.BranchOwner, repo)
			if err != nil {
				return nil, errors.Wrapf(err, "fail to get clone token for %s/%s", gitOptions.BranchOwner, repo)
			}

			cloneTokenSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      strings.ReplaceAll(strings.ToLower(fmt.Sprintf("%s-%s", repo, namespace)), "_", ""),
					Namespace: "kpack",
					Annotations: map[string]string{
						"kpack.io/git": "https://github.com",
					},
				},
				StringData: map[string]string{
					"username": "x-access-token",
					"password": cloneToken,
				},
				Type: "kubernetes.io/basic-auth",
			}

			err = b.clusterClient.CreateSecret(ctx, cloneTokenSecret)
			if err != nil {
				return nil, errors.Wrap(err, "fail to add github token secret into cluster")
			}

			cloneTokenSecretName = pointer.String(cloneTokenSecret.GetName())
			cloneTokenSecrets[repo] = cloneTokenSecretName
		}

		branch := gitOptions.Branch
		branchExists, err := b.gitClient.DoesBranchExist(ctx, gitOptions.Owner, repo, branch, gitOptions.BranchOwner)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to check if branch %s for repo %s/%s exists", branch, gitOptions.BranchOwner, repo)
		}
		if !branchExists {
			defaultBranch, err := b.gitClient.GetDefaultBranch(ctx, gitOptions.Owner, repo, gitOptions.BranchOwner)
			if err != nil {
				return nil, errors.Wrapf(err, "fail to get default branch for repo %s/%s", gitOptions.BranchOwner, repo)
			}
			branch = defaultBranch
		}

		svcAcc := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      service.ID,
				Namespace: "kpack",
			},
			Secrets:          []corev1.ObjectReference{{Name: "kpack-registry-credentials"}},
			ImagePullSecrets: []corev1.LocalObjectReference{{Name: "kpack-registry-credentials"}},
		}
		if cloneTokenSecretName != nil {
			svcAcc.Secrets = append(svcAcc.Secrets, corev1.ObjectReference{Name: *cloneTokenSecretName})
		}

		err = b.clusterClient.CreateServiceAccount(ctx, svcAcc)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to create service account to build service %s", service.ID)
		}

		vars, err := b.envVarsProvider.ListByRepoBranch(ctx, gitOptions.Owner, repo, branch)
		if err != nil {
			return nil, errors.Wrap(err, "fail to list env vars by repo")
		}

		addedVariables := make(map[string]struct{})
		envs := []corev1.EnvVar{}
		for _, v := range vars {
			envs = append(envs, corev1.EnvVar{
				Name:  v.Name,
				Value: v.Value,
				// TODO: use ValueFrom and store the vars in a secret
			})
			addedVariables[v.Name] = struct{}{}
		}

		for k, v := range service.Env {
			if _, ok := addedVariables[k]; ok {
				continue
			}

			envs = append(envs, corev1.EnvVar{Name: k, Value: v})
		}

		labels := getLabels(service.ID, serviceName, gitOptions)
		build := &kpackBuild.Build{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Build",
				APIVersion: "kpack.io/v1alpha2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        service.ID,
				Namespace:   "kpack",
				Labels:      labels,
				Annotations: labels,
			},
			Spec: kpackBuild.BuildSpec{
				Builder: kpackCore.BuildBuilderSpec{
					Image: "ergomake/kpack-builder",
				},
				RunImage: kpackBuild.BuildSpecImage{
					Image: "paketobuildpacks/run-jammy-base",
				},
				ServiceAccountName: svcAcc.GetName(),
				Source: kpackCore.SourceConfig{
					Git: &kpackCore.Git{
						URL:      fmt.Sprintf("https://github.com/%s/%s", gitOptions.BranchOwner, repo),
						Revision: branch,
					},
					SubPath: buildPath,
				},
				Tags: []string{service.Image},
				Env:  envs,
				Tolerations: []corev1.Toleration{
					{
						Key:      "preview.ergomake.dev/domain",
						Operator: corev1.TolerationOpEqual,
						Value:    "build",
						Effect:   corev1.TaintEffectNoSchedule,
					},
				},
				NodeSelector: map[string]string{
					"preview.ergomake.dev/role": "build",
				},
			},
		}

		builds = append(builds, build)
	}

	err := b.clusterClient.ApplyKPackBuilds(ctx, builds)
	if err != nil {
		return nil, errors.Wrap(err, "fail to apply kpack build")
	}

	return &BuildImagesResult{}, nil
}

func (b *builder) buildImagesWithKanikoFromGit(ctx context.Context, namespace string, gitOptions GitOptions) (*BuildImagesResult, error) {
	cloneTokenSecrets := make(map[string]*string)
	jobs := []*batchv1.Job{}
	for k, service := range b.komposeObject.ServiceConfigs {
		if service.Build == "" && service.Dockerfile == "" {
			continue
		}

		repo, buildPath := ComputeRepoAndBuildPath(b.projectPath, b.configFilePath, service.Build, gitOptions.Repo)

		cloneTokenSecretName, ok := cloneTokenSecrets[repo]
		if !ok && !gitOptions.IsPublic {
			cloneToken, err := b.gitClient.GetCloneToken(ctx, gitOptions.BranchOwner, repo)
			if err != nil {
				return nil, errors.Wrapf(err, "fail to get clone token for %s/%s", gitOptions.BranchOwner, repo)
			}

			cloneTokenSecret := makeCloneTokenSecret(namespace, repo, cloneToken)
			err = b.clusterClient.CreateSecret(ctx, cloneTokenSecret)
			if err != nil {
				return nil, errors.Wrap(err, "fail to add github token secret into cluster")
			}

			cloneTokenSecretName = pointer.String(cloneTokenSecret.GetName())
			cloneTokenSecrets[repo] = cloneTokenSecretName
		}

		branch := gitOptions.Branch
		branchExists, err := b.gitClient.DoesBranchExist(ctx, gitOptions.Owner, repo, branch, gitOptions.BranchOwner)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to check if branch %s for repo %s/%s exists", branch, gitOptions.BranchOwner, repo)
		}
		if !branchExists {
			defaultBranch, err := b.gitClient.GetDefaultBranch(ctx, gitOptions.Owner, repo, gitOptions.BranchOwner)
			if err != nil {
				return nil, errors.Wrapf(err, "fail to get default branch for repo %s/%s", gitOptions.BranchOwner, repo)
			}
			branch = defaultBranch
		}

		vars, err := b.envVarsProvider.ListByRepoBranch(ctx, gitOptions.Owner, repo, branch)
		if err != nil {
			return nil, errors.Wrap(err, "fail to list env vars by repo")
		}

		spec := b.makeJobSpec(
			b.environment.Services[k].ID,
			k,
			service,
			buildPath,
			vars,
			gitOptions,
			"dir:///workspace",
		)
		spec.Spec.Template.Spec.InitContainers = []corev1.Container{
			b.makeInitContainer(spec, gitOptions.BranchOwner, repo, branch, cloneTokenSecretName),
		}

		job, err := b.clusterClient.CreateJob(ctx, spec)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to create build job for service %s", k)
		}
		jobs = append(jobs, job)
	}

	jobCtx, cancelFn := context.WithTimeout(ctx, time.Hour)
	defer cancelFn()
	result, err := b.clusterClient.WaitJobs(jobCtx, jobs)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to wait for build jobs to complete")
	}

	return &BuildImagesResult{result.Failed}, nil
}

func makeCloneTokenSecret(namespace, repo, token string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.ReplaceAll(strings.ToLower(fmt.Sprintf("%s-%s", repo, namespace)), "_", ""),
			Namespace: "preview-builds",
		},
		Data: map[string][]byte{
			"token": []byte(token),
		},
	}
}

func (b *builder) makeJobSpec(
	serviceID string,
	serviceName string,
	service kobject.ServiceConfig,
	buildPath string,
	vars []envvars.EnvVar,
	gitOptions GitOptions,
	context string,
) *batchv1.Job {

	buildArgsSet := make(map[string]struct{})

	buildArgs := []string{}
	for k, v := range service.BuildArgs {
		if v == nil {
			continue
		}

		buildArgs = append(buildArgs, "--build-arg", fmt.Sprintf("%s=%s", k, *v))
		buildArgsSet[k] = struct{}{}
	}

	for _, v := range vars {
		if _, ok := buildArgsSet[v.Name]; ok {
			continue
		}

		buildArgs = append(buildArgs, "--build-arg", fmt.Sprintf("%s=%s", v.Name, v.Value))
	}

	args := append([]string{
		fmt.Sprintf("--context=%s", context),
		fmt.Sprintf("--dockerfile=%s", service.Dockerfile),
		fmt.Sprintf("--context-sub-path=%s", buildPath),
		"--destination=" + service.Image,
		"--use-new-run",
		"--cleanup",
		"--snapshot-mode=redo",
	}, buildArgs...)

	if insecureRegistry != "" {
		// get the hostname and port from Image using stdlib
		args = append(args, fmt.Sprintf("--insecure-registry=%s", insecureRegistry))
	}

	labels := getLabels(serviceID, serviceName, gitOptions)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        serviceID,
			Namespace:   "preview-builds",
			Labels:      labels,
			Annotations: labels,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: int32Ptr(120),
			ActiveDeadlineSeconds:   int64Ptr(30 * 60),

			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "preview-builds",
					Labels:      labels,
					Annotations: labels,
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{{Name: b.dockerhubPullSecretName}},
					Containers: []corev1.Container{
						{
							Name:  serviceID,
							Image: "gcr.io/kaniko-project/executor:latest",
							Args:  args,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workspace",
									MountPath: "/workspace",
								},
							},
							ImagePullPolicy: "IfNotPresent",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("7Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("7Gi"),
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "AWS_REGION",
									Value: "us-east-1",
								},
								{
									Name:  "AWS_ACCESS_KEY_ID",
									Value: "nop",
								},
								{
									Name:  "AWS_SECRET_ACCESS_KEY",
									Value: "nop",
								},
							},
						},
					},
					ServiceAccountName: "preview-builder",
					RestartPolicy:      corev1.RestartPolicyNever,
					Volumes: []corev1.Volume{
						{
							Name: "workspace",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      "preview.ergomake.dev/domain",
							Operator: corev1.TolerationOpEqual,
							Value:    "build",
							Effect:   corev1.TaintEffectNoSchedule,
						},
					},
					NodeSelector: map[string]string{
						"preview.ergomake.dev/role": "build",
					},
				},
			},
			BackoffLimit: int32Ptr(0),
		},
	}

	if os.Getenv("CLUSTER") == "eks" {
		appendUserlandCreds(job)
	}

	return job
}

func (b *builder) makeInitContainer(
	jobSpec *batchv1.Job,
	githubOwner string,
	githubRepo string,
	githubBranch string,
	githubTokenSecretName *string,
) corev1.Container {
	cmd := append([]string{
		"git",
		"clone",
		b.gitClient.GetCloneUrl(),
	}, b.gitClient.GetCloneParams()...)

	cmd = append(cmd, "/workspace")

	env := []corev1.EnvVar{
		{
			Name:  "OWNER",
			Value: githubOwner,
		},
		{
			Name:  "REPO",
			Value: githubRepo,
		},
		{
			Name:  "BRANCH",
			Value: githubBranch,
		},
	}

	if githubTokenSecretName != nil {
		env = append(env, corev1.EnvVar{
			Name: "GIT_TOKEN",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: "token",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: *githubTokenSecretName,
					},
				},
			},
		})
	}

	return corev1.Container{
		Name:            "git-clone",
		Image:           "alpine/git",
		Command:         cmd,
		ImagePullPolicy: "IfNotPresent",
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "workspace",
				MountPath: "/workspace",
			},
		},
		Env: env,
	}
}

func appendUserlandCreds(job *batchv1.Job) {
	dockerConfigVolumeMount := corev1.VolumeMount{
		Name:      "docker-config",
		MountPath: "/kaniko/.docker/",
	}
	job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, dockerConfigVolumeMount)

	dockerConfigVolume := corev1.Volume{
		Name: "docker-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "docker-config",
				},
			},
		},
	}

	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, dockerConfigVolume)
}

func getLabels(serviceID, serviceName string, gitOptions GitOptions) map[string]string {
	return map[string]string{
		"app":                              serviceID,
		"preview.ergomake.dev/id":          serviceID,
		"preview.ergomake.dev/service":     serviceName,
		"preview.ergomake.dev/owner":       gitOptions.Owner,
		"preview.ergomake.dev/BranchOwner": gitOptions.BranchOwner,
		"preview.ergomake.dev/repo":        gitOptions.Repo,
		"preview.ergomake.dev/sha":         gitOptions.SHA,
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}
