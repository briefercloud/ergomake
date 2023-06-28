package transformer

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

	"github.com/ergomake/ergomake/internal/envvars"
)

type BuildImagesResult struct {
	FailedJobs []*batchv1.Job
}

func (bir *BuildImagesResult) Failed() bool {
	return len(bir.FailedJobs) > 0
}

func (c *gitCompose) computeRepoAndBuildPath(buildPath string, defaultRepo string) (string, string) {
	projectPath := filepath.Clean(c.projectPath)
	fullBuildPath := filepath.Clean(path.Join(path.Dir(c.composePath), buildPath))

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

func (c *gitCompose) buildImages(
	ctx context.Context,
	namespace string,
) (*BuildImagesResult, error) {

	cloneTokenSecrets := make(map[string]*string)
	jobs := []*batchv1.Job{}
	for k, service := range c.komposeObject.ServiceConfigs {
		if service.Build == "" && service.Dockerfile == "" {
			continue
		}

		repo, buildPath := c.computeRepoAndBuildPath(service.Build, c.repo)

		cloneTokenSecretName, ok := cloneTokenSecrets[repo]
		if !ok && c.needsToken {
			cloneToken, err := c.gitClient.GetCloneToken(ctx, c.branchOwner, repo)
			if err != nil {
				return nil, errors.Wrapf(err, "fail to get clone token for %s/%s", c.branchOwner, repo)
			}

			cloneTokenSecret := makeCloneTokenSecret(namespace, repo, cloneToken)
			err = c.clusterClient.CreateSecret(ctx, cloneTokenSecret)
			if err != nil {
				return nil, errors.Wrap(err, "fail to add github token secret into cluster")
			}

			cloneTokenSecretName = pointer.String(cloneTokenSecret.GetName())
			cloneTokenSecrets[repo] = cloneTokenSecretName
		}

		branch := c.branch
		branchExists, err := c.gitClient.DoesBranchExist(ctx, c.branchOwner, repo, branch)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to check if branch %s for repo %s/%s exists", branch, c.branchOwner, repo)
		}
		if !branchExists {
			defaultBranch, err := c.gitClient.GetDefaultBranch(ctx, c.branchOwner, repo)
			if err != nil {
				return nil, errors.Wrapf(err, "fail to get default branch for repo %s/%s", c.branchOwner, repo)
			}
			branch = defaultBranch
		}

		vars, err := c.envVarsProvider.ListByRepo(ctx, c.owner, repo)
		if err != nil {
			return nil, errors.Wrap(err, "fail to list env vars by repo")
		}

		spec := c.makeJobSpec(c.compose.Services[k].ID, k, service, buildPath, vars)
		spec.Spec.Template.Spec.InitContainers = []corev1.Container{
			c.makeInitContainer(spec, c.branchOwner, repo, branch, cloneTokenSecretName),
		}

		job, err := c.clusterClient.CreateJob(ctx, spec)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to create build job for service %s", k)
		}
		jobs = append(jobs, job)
	}

	jobCtx, cancelFn := context.WithTimeout(ctx, time.Hour)
	defer cancelFn()
	result, err := c.clusterClient.WaitJobs(jobCtx, jobs)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to wait for build jobs to complete")
	}

	return &BuildImagesResult{result.Failed}, nil
}

func makeCloneTokenSecret(namespace, repo, token string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.ToLower(fmt.Sprintf("%s-%s", repo, namespace)),
			Namespace: "preview-builds",
		},
		Data: map[string][]byte{
			"token": []byte(token),
		},
	}
}

func (c *gitCompose) makeJobSpec(serviceID string, serviceName string, service kobject.ServiceConfig, buildPath string, vars []envvars.EnvVar) *batchv1.Job {

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
		"--context=dir:///workspace",
		fmt.Sprintf("--dockerfile=%s", service.Dockerfile),
		fmt.Sprintf("--context-sub-path=%s", buildPath),
		"--destination=" + service.Image,
		"--use-new-run",
		"--compressed-caching=false",
		"--cleanup",
		"--snapshot-mode=redo",
		"--cache",
		fmt.Sprintf("--cache-repo=%s", userlandRegistry),
	}, buildArgs...)

	if insecureRegistry != "" {
		// get the hostname and port from Image using stdlib
		args = append(args, fmt.Sprintf("--insecure-registry=%s", insecureRegistry))
	}

	labels := map[string]string{
		"app":                              serviceID,
		"preview.ergomake.dev/id":          serviceID,
		"preview.ergomake.dev/service":     serviceName,
		"preview.ergomake.dev/owner":       c.owner,
		"preview.ergomake.dev/branchOwner": c.branchOwner,
		"preview.ergomake.dev/repo":        c.repo,
		"preview.ergomake.dev/sha":         c.sha,
	}
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

func (c *gitCompose) makeInitContainer(
	jobSpec *batchv1.Job,
	githubOwner string,
	githubRepo string,
	githubBranch string,
	githubTokenSecretName *string,
) corev1.Container {
	cmd := append([]string{
		"git",
		"clone",
		c.gitClient.GetCloneUrl(),
	}, c.gitClient.GetCloneParams()...)

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
		Name:    "git-clone",
		Image:   "alpine/git",
		Command: cmd,
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

func int32Ptr(i int32) *int32 {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}
