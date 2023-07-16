package git

import (
	"context"
	"encoding/base64"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"github.com/cbroglie/mustache"
	"github.com/google/uuid"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/loader"
	"github.com/kubernetes/kompose/pkg/transformer/kubernetes"
	"github.com/pkg/errors"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/envvars"
	"github.com/ergomake/ergomake/internal/ergopack"
	"github.com/ergomake/ergomake/internal/git"
	"github.com/ergomake/ergomake/internal/logger"
	"github.com/ergomake/ergomake/internal/privregistry"
	"github.com/ergomake/ergomake/internal/transformer"
	"github.com/ergomake/ergomake/internal/transformer/builder"
)

var clusterDomain string
var userlandRegistry string

func init() {
	setUserlandRegistry()
	setDomain()
}

func setUserlandRegistry() {
	cluster := os.Getenv("CLUSTER")
	if cluster != "eks" {
		userlandRegistry = "host.minikube.internal:5001/library"
		return
	}

	userlandRegistry = os.Getenv("ECR_USERLAND_REPO")
	if userlandRegistry == "" {
		logger.Get().Fatal().Msg("ECR_USERLAND_REPO environment variable not set")
	}
}

func setDomain() {
	cluster := os.Getenv("CLUSTER")
	clusterDomain = os.Getenv("CLUSTER_DOMAIN")
	if clusterDomain == "" {
		if cluster == "eks" {
			logger.Get().Fatal().Msg("CLUSTER_DOMAIN environment variable not set")
			return
		}

		clusterDomain = "env.ergomake.test"
	}
}

type gitCompose struct {
	clusterClient           cluster.Client
	gitClient               git.RemoteGitClient
	db                      *database.DB
	envVarsProvider         envvars.EnvVarsProvider
	privRegistryProvider    privregistry.PrivRegistryProvider
	dockerhubPullSecretName string
	s3Bucket                string
	awsAccessKey            string
	awsSecretAccessKey      string
	gitOptions              builder.GitOptions

	prepared bool

	projectPath    string
	configFilePath string
	dbEnvironment  *database.Environment
	environment    *transformer.Environment
	isCompose      bool
	komposeObject  *kobject.KomposeObject
	cleanup        func()
}

func NewGitCompose(
	clusterClient cluster.Client,
	gitClient git.RemoteGitClient,
	db *database.DB,
	envVarsProvider envvars.EnvVarsProvider,
	privRegistryProvider privregistry.PrivRegistryProvider,
	s3Bucket string,
	awsAccessKey string,
	awsSecretAccessKey string,
	gitOptions builder.GitOptions,
	dockerhubPullSecretName string,
) *gitCompose {
	return &gitCompose{
		clusterClient:           clusterClient,
		gitClient:               gitClient,
		db:                      db,
		envVarsProvider:         envVarsProvider,
		privRegistryProvider:    privRegistryProvider,
		s3Bucket:                s3Bucket,
		awsAccessKey:            awsAccessKey,
		awsSecretAccessKey:      awsSecretAccessKey,
		gitOptions:              gitOptions,
		dockerhubPullSecretName: dockerhubPullSecretName,
	}
}

func (c *gitCompose) Prepare(ctx context.Context, id uuid.UUID) (*transformer.PrepareResult, error) {
	namespace := id.String()
	dbEnv := database.NewEnvironment(
		id,
		c.gitOptions.Owner,
		c.gitOptions.BranchOwner,
		c.gitOptions.Repo,
		c.gitOptions.Branch,
		c.gitOptions.PrNumber,
		c.gitOptions.Author,
		database.EnvPending,
	)
	err := c.db.Create(&dbEnv).Error
	if err != nil {
		return nil, errors.Wrap(err, "fail to create environment in db")
	}

	c.dbEnvironment = dbEnv

	loadErgopackResult, err := c.loadErgopack(ctx, namespace)
	if err != nil {
		return nil, c.fail(errors.Wrap(err, "fail to load ergopack"))
	}

	c.prepared = true

	if loadErgopackResult.Skip {
		err := c.db.Delete(c.dbEnvironment).Error
		if err != nil {
			logger.Ctx(ctx).Err(err).Msg("fail to delete skipped environment")
		}
	}

	if loadErgopackResult.ValidationError != nil {
		dbEnv.Status = database.EnvDegraded
		dbEnv.DegradedReason, err = json.Marshal(loadErgopackResult.ValidationError)
		if err != nil {
			return nil, errors.Wrap(err, "fail to marshal validation error")
		}

		err = c.db.Save(&dbEnv).Error
		if err != nil {
			return nil, errors.Wrap(err, "fail to save degraded reason to db")
		}

		return &transformer.PrepareResult{
			Environment:     dbEnv,
			Skip:            false,
			ValidationError: loadErgopackResult.ValidationError,
		}, nil
	}

	return &transformer.PrepareResult{
		Environment: dbEnv,
		Skip:        loadErgopackResult.Skip,
	}, nil
}

func (c *gitCompose) Cleanup() {
	if c.cleanup != nil {
		c.cleanup()
	}
}

func (c *gitCompose) fail(origErr error) error {
	err := c.db.Model(&c.dbEnvironment).Update("status", database.EnvDegraded).Error
	if err != nil {
		return errors.Wrap(stderrors.Join(origErr, err), "fail to update db environment status to degraded")
	}

	return origErr
}

func (c *gitCompose) Transform(ctx context.Context, id uuid.UUID) (*transformer.TransformResult, error) {
	if !c.prepared {
		return nil, errors.New("called Transform before calling Prepare")
	}

	namespace := id.String()
	result := &transformer.TransformResult{IsCompose: c.isCompose}

	err := c.saveServices(ctx, id, c.environment)
	if err != nil {
		return nil, c.fail(errors.Wrap(err, "fail to save services"))
	}

	builder := builder.NewBuilder(
		c.clusterClient,
		c.gitClient,
		c.envVarsProvider,
		c.db,
		c.s3Bucket,
		c.awsAccessKey,
		c.awsSecretAccessKey,
		c.projectPath,
		c.configFilePath,
		c.dbEnvironment,
		c.isCompose,
		c.komposeObject,
		c.environment,
		c.dockerhubPullSecretName,
	)
	buildImagesRes, err := builder.BuildImagesFromGit(ctx, namespace, c.gitOptions)
	if err != nil {
		return nil, c.fail(errors.Wrap(err, "fail to build images"))
	}

	if buildImagesRes.Failed() {
		result.FailedJobs = buildImagesRes.FailedJobs
		return result, c.fail(nil)
	}

	var objects []runtime.Object
	if c.isCompose {
		objs, err := c.transformCompose(ctx, namespace)
		if err != nil {
			return nil, c.fail(errors.Wrap(err, "fail to tranform compose into k8s objects"))
		}
		objects = objs
		err = c.db.Model(&c.dbEnvironment).Update("status", database.EnvSuccess).Error
		if err != nil {
			return nil, errors.Wrap(err, "fail to update environment status to success in db")
		}
	} else {
		objs, err := c.makeClusterObjects(ctx, namespace)
		if err != nil {
			return nil, c.fail(errors.Wrap(err, "fail to make cluster objects"))
		}

		objects = objs
	}

	result.ClusterEnv = &cluster.ClusterEnv{
		Namespace: namespace,
		Objects:   objects,
	}
	result.Environment = c.environment

	return result, nil
}

func (c *gitCompose) makeClusterObjects(ctx context.Context, namespace string) ([]runtime.Object, error) {
	vars, err := c.envVarsProvider.ListByRepoBranch(ctx, c.gitOptions.Owner, c.gitOptions.Repo, c.gitOptions.Branch)
	if err != nil {
		return nil, errors.Wrap(err, "fail to list env vars by repo")
	}

	secret, err := c.clusterClient.CopySecret(ctx, "kpack", namespace, "kpack-registry-credentials")
	if err != nil {
		return nil, errors.Wrap(err, "fail to copy registry credentials")
	}

	objs := []runtime.Object{secret}

	envVarsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "env-vars",
			Namespace: namespace,
		},
		StringData: map[string]string{},
	}
	env := []corev1.EnvVar{}

	dbVars := make(map[string]struct{})
	for _, v := range vars {
		envVarsSecret.StringData[v.Name] = v.Value

		env = append(env, corev1.EnvVar{
			Name: v.Name,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "env-vars",
					},
					Key: v.Name,
				},
			},
		})
		dbVars[v.Name] = struct{}{}
	}

	objs = append(objs, envVarsSecret)

	for serviceName, envService := range c.environment.Services {
		for k, v := range envService.Env {
			if _, ok := dbVars[k]; ok {
				continue
			}

			env = append(env, corev1.EnvVar{Name: k, Value: v})
		}

		labels := c.getLabels(envService.ID, serviceName)

		containerPorts := []corev1.ContainerPort{}
		servicePorts := []corev1.ServicePort{}
		ports := map[int]struct{}{}
		for _, strPort := range append(envService.InternalPorts, envService.PublicPort) {
			if strPort == "" {
				continue
			}

			port, err := strconv.Atoi(strPort)
			if err != nil {
				logger.Ctx(ctx).Warn().AnErr("err", err).
					Str("strPort", strPort).
					Msg("fail to convert string port to int")
				continue
			}

			if _, ok := ports[port]; ok {
				continue
			}

			containerPorts = append(containerPorts, corev1.ContainerPort{
				ContainerPort: int32(port),
			})
			servicePorts = append(servicePorts, corev1.ServicePort{
				Port: int32(port),
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: int32(port),
				},
			})
		}

		container := corev1.Container{
			Name:  serviceName,
			Image: envService.Image,
			Ports: containerPorts,
			Env:   env,
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceEphemeralStorage: resource.MustParse("2Gi"),
					corev1.ResourceMemory:           resource.MustParse("1Gi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceEphemeralStorage: resource.MustParse("2Gi"),
					corev1.ResourceMemory:           resource.MustParse("1Gi"),
				},
			},
			ImagePullPolicy: "IfNotPresent",
		}

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:        serviceName,
				Namespace:   namespace,
				Labels:      labels,
				Annotations: labels,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: pointer.Int32(0),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"preview.ergomake.dev/service": serviceName,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      labels,
						Annotations: labels,
					},
					Spec: corev1.PodSpec{
						ImagePullSecrets: []corev1.LocalObjectReference{{Name: secret.GetName()}},
						Containers:       []corev1.Container{container},
						NodeSelector: map[string]string{
							"preview.ergomake.dev/role": "preview",
						},
						SecurityContext: &corev1.PodSecurityContext{
							SeccompProfile: &corev1.SeccompProfile{
								Type: "RuntimeDefault",
							},
						},
						Tolerations: []corev1.Toleration{
							{
								Key:      "preview.ergomake.dev/domain",
								Operator: "Equal",
								Value:    "previews",
								Effect:   "NoSchedule",
							},
						},
					},
				},
			},
		}

		objs = append(objs, deployment)

		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        serviceName,
				Namespace:   namespace,
				Labels:      labels,
				Annotations: labels,
			},
			Spec: corev1.ServiceSpec{
				Ports: servicePorts,
				Selector: map[string]string{
					"preview.ergomake.dev/service": serviceName,
				},
			},
		}
		objs = append(objs, service)

		if envService.PublicPort != "" {
			port, err := strconv.Atoi(envService.PublicPort)
			if err != nil {
				logger.Ctx(ctx).Warn().AnErr("err", err).Str("strPort", envService.PublicPort).
					Msg("fail to convert PublicPort to int")
				continue
			}

			pathType := networkingv1.PathTypePrefix
			ingress := &networkingv1.Ingress{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:        serviceName,
					Namespace:   namespace,
					Labels:      labels,
					Annotations: labels,
				},
				Spec: networkingv1.IngressSpec{
					IngressClassName: pointer.String("nginx"),
					Rules: []networkingv1.IngressRule{
						{
							Host: envService.Url,
							IngressRuleValue: networkingv1.IngressRuleValue{
								HTTP: &networkingv1.HTTPIngressRuleValue{
									Paths: []networkingv1.HTTPIngressPath{
										{
											Path:     "/",
											PathType: &pathType,
											Backend: networkingv1.IngressBackend{
												Service: &networkingv1.IngressServiceBackend{
													Name: serviceName,
													Port: networkingv1.ServiceBackendPort{
														Number: int32(port),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			objs = append(objs, ingress)
		}
	}

	return objs, nil
}

func (c *gitCompose) saveServices(ctx context.Context, envID uuid.UUID, compose *transformer.Environment) error {
	var services []database.Service
	for name, service := range compose.Services {
		buildStatus := "image"
		if service.Build != "" {
			buildStatus = "building"
		}

		services = append(services, database.Service{
			ID:            service.ID,
			Name:          name,
			EnvironmentID: envID,
			Url:           service.Url,
			Build:         service.Build,
			BuildStatus:   buildStatus,
			Image:         service.Image,
			Index:         service.Index,
			PublicPort:    service.PublicPort,
			InternalPorts: service.InternalPorts,
		})
	}

	if len(services) == 0 {
		return nil
	}

	return c.db.Create(&services).Error
}

// returns empty when service should not be exposed
func (c *gitCompose) getUrl(service kobject.ServiceConfig) string {
	for _, port := range service.Port {
		suffix := c.gitOptions.Branch
		if c.gitOptions.PrNumber != nil {
			suffix = strconv.Itoa(*c.gitOptions.PrNumber)
		}

		if port.HostPort > 0 {
			return strings.ToLower(fmt.Sprintf(
				"%s-%s-%s-%s.%s",
				service.Name,
				c.gitOptions.Owner,
				strings.ReplaceAll(c.gitOptions.Repo, "_", ""),
				suffix,
				clusterDomain,
			))
		}
	}

	return ""
}

func (c *gitCompose) fixComposeObject(projectPath, namespace string) error {
	for k, service := range c.komposeObject.ServiceConfigs {
		if service.Build != "" {
			service.Image = fmt.Sprintf(
				"%s:%s-%s",
				userlandRegistry,
				namespace,
				service.Name,
			)
			service.Build = strings.Replace(service.Build, projectPath, "", 1)
		}

		service.ExposeService = c.getUrl(service)
		if service.ExposeService != "" {
			service.ExposeServiceIngressClassName = "nginx"
		}

		err := evaluateLabels(&service, c.environment)
		if err != nil {
			return errors.Wrap(err, "fail to evaluate ergomake specific labels")
		}

		c.removeUnsupportedVolumes(&service)

		c.komposeObject.ServiceConfigs[k] = service
	}

	return nil
}

func (c *gitCompose) removeUnsupportedVolumes(service *kobject.ServiceConfig) {
	volumes := []kobject.Volumes{}
	for _, vol := range service.Volumes {
		if strings.HasPrefix(vol.MountPath, ":") {
			continue
		}
		volumes = append(volumes, vol)
	}

	service.Volumes = volumes
}

type LoadErgopackResult struct {
	Skip            bool
	ValidationError *transformer.ProjectValidationError
}

func (c *gitCompose) loadErgopack(ctx context.Context, namespace string) (*LoadErgopackResult, error) {
	projectPath, err := c.cloneRepo(ctx, namespace)
	if err != nil {
		return nil, errors.Wrap(err, "fail to clone repo from github")
	}

	c.projectPath = projectPath

	c.cleanup = func() {
		err := os.RemoveAll(projectPath)
		if err != nil {
			logger.Ctx(ctx).Err(err).Str("projectPath", projectPath).Str("namespace", namespace).
				Msg("fail to cleanup project path")
		}
	}

	_, err = os.Stat(path.Join(c.projectPath, ".ergomake"))
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}

		return &LoadErgopackResult{Skip: true}, errors.Wrap(err, "fail to check if .ergomake folder exists")
	}

	validationRes, err := transformer.Validate(projectPath)
	if err != nil {
		return nil, errors.Wrap(err, "fail to validate project")
	}

	if validationRes.ProjectValidationError != nil {
		return &LoadErgopackResult{Skip: false, ValidationError: validationRes.ProjectValidationError}, nil
	}

	c.configFilePath = validationRes.ConfigFilePath
	c.isCompose = validationRes.IsCompose

	if c.isCompose {
		c.dbEnvironment.BuildTool = "kaniko"
	} else {
		c.dbEnvironment.BuildTool = "buildpacks"
	}

	err = c.db.Save(&c.dbEnvironment).Error
	if err != nil {
		return nil, errors.Wrap(err, "fail to save env build_tool to db")
	}

	configBytes, err := ioutil.ReadFile(c.configFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to read compose at %s", c.configFilePath)
	}
	configStr := string(configBytes)

	if c.isCompose {
		loader, err := loader.GetLoader("compose")
		if err != nil {
			return nil, errors.Wrap(err, "fail to get kompose loader")
		}

		komposeObject, err := loader.LoadFile([]string{c.configFilePath})
		if err != nil {
			return nil, errors.Wrapf(err, "fail to load compose %s", c.configFilePath)
		}
		c.komposeObject = &komposeObject

		c.environment = c.makeEnvironmentFromKObjectServices(
			komposeObject.ServiceConfigs,
			configStr,
		)

		err = c.fixComposeObject(projectPath, namespace)
		if err != nil {
			return nil, errors.Wrap(err, "fail to fix compose object")
		}
	} else {
		var pack ergopack.Ergopack
		err := yaml.Unmarshal(configBytes, &pack)
		if err != nil {
			relativePath, err := filepath.Rel(projectPath, c.configFilePath)
			if err != nil {
				relativePath = c.configFilePath
			}

			return &LoadErgopackResult{
				Skip: false,
				ValidationError: &transformer.ProjectValidationError{
					T:       "invalid-ergopack",
					Message: fmt.Sprintf("Ergopack file has syntax error\n```\n%s: %s\n```", relativePath, err.Error()),
				},
			}, nil
		}

		c.environment = c.makeEnvironmentFromErgopack(ctx, &pack, string(configBytes))
	}

	return &LoadErgopackResult{}, nil
}

func (c *gitCompose) transformCompose(ctx context.Context, namespace string) ([]runtime.Object, error) {
	// Create the options for the conversion to Kubernetes objects.
	convertOptions := kobject.ConvertOptions{
		ToStdout:   true,
		CreateD:    true,
		Replicas:   1,
		PushImage:  false,
		InputFiles: []string{c.configFilePath},
		Volumes:    "configMap",
		Controller: "deployment",
	}

	// Get the Kubernetes transformer.
	transformer := &kubernetes.Kubernetes{
		Opt: convertOptions,
	}

	// Transform the Docker Compose objects into Kubernetes objects.
	objects, err := transformer.Transform(*c.komposeObject, convertOptions)
	if err != nil {
		return nil, errors.Wrap(err, "fail to tranform compose into k8s objects")
	}

	extraObjs, err := c.fixOutput(ctx, &objects, namespace)

	return append(objects, extraObjs...), errors.Wrap(err, "fail to fix output")
}

func (c *gitCompose) cloneRepo(ctx context.Context, namespace string) (string, error) {
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("ergomake-%s-%s-%s", c.gitOptions.Owner, c.gitOptions.Repo, namespace))
	if err != nil {
		return "", errors.Wrap(err, "fail to make temp dir")
	}

	// it is important that the folder name is not too big that's why we create yet another dir
	dir := path.Join(tmpDir, c.gitOptions.Repo)
	err = os.Mkdir(dir, 0700)
	if err != nil {
		return "", errors.Wrap(err, "fail to make inner dir inside temp dir")
	}

	err = c.gitClient.CloneRepo(ctx, c.gitOptions.BranchOwner, c.gitOptions.Repo, c.gitOptions.Branch, dir, c.gitOptions.IsPublic)

	return dir, errors.Wrap(err, "fail to clone from github")
}

func (c *gitCompose) makeEnvironmentFromKObjectServices(komposeServices map[string]kobject.ServiceConfig, rawCompose string) *transformer.Environment {
	services := map[string]transformer.EnvironmentService{}
	for _, service := range komposeServices {
		services[service.Name] = transformer.EnvironmentService{
			ID:    uuid.NewString(),
			Url:   c.getUrl(service),
			Image: service.Image,
			Build: service.Build,
		}
	}

	return transformer.NewEnvironment(services, rawCompose)
}

func (c *gitCompose) makeEnvironmentFromErgopack(ctx context.Context, pack *ergopack.Ergopack, rawFile string) *transformer.Environment {
	services := map[string]transformer.EnvironmentService{}
	i := 0
	for name, service := range pack.Apps {
		url := ""
		if service.PublicPort != "" {
			suffix := c.gitOptions.Branch
			if c.gitOptions.PrNumber != nil {
				suffix = strconv.Itoa(*c.gitOptions.PrNumber)
			}

			url = strings.ToLower(fmt.Sprintf(
				"%s-%s-%s-%s.%s",
				name,
				c.gitOptions.Owner,
				strings.ReplaceAll(c.gitOptions.Repo, "_", ""),
				suffix,
				clusterDomain,
			))
		}

		id := uuid.NewString()
		image := service.Image
		if image == "" {
			image = strings.ToLower(fmt.Sprintf("ergomake/%s-%s-%s:%s", c.gitOptions.Owner, c.gitOptions.Repo, name, id))
		}

		services[name] = transformer.EnvironmentService{
			ID:            id,
			Url:           url,
			Image:         image,
			Build:         service.Path,
			PublicPort:    service.PublicPort,
			InternalPorts: service.InternalPorts,
			Index:         i,
			Env:           service.Env,
		}
		i += 1
	}

	env := transformer.NewEnvironment(services, rawFile)

	mustache.AllowMissingVariables = false
	templateContext := env.ToMap()

	for name, service := range env.Services {
		for k, v := range service.Env {
			newV, err := mustache.Render(v, templateContext)
			if err != nil {
				logger.Ctx(ctx).Err(err).Str("var", k).Msg("fail to render env var")
			}

			service.Env[k] = newV
		}

		env.Services[name] = service
	}

	return env
}

func evaluateLabels(service *kobject.ServiceConfig, env *transformer.Environment) error {
	mustache.AllowMissingVariables = false

	templateContext := env.ToMap()

	for label, value := range service.Labels {
		replaceArgLabel := "dev.ergomake.env.replace-arg."

		if strings.HasPrefix(label, replaceArgLabel) {
			varName := strings.TrimPrefix(label, replaceArgLabel)
			replacedValue, err := mustache.Render(value, templateContext)
			if err != nil {
				return errors.Wrapf(
					err,
					"fail to render mustache template for replace-arg label var=%s value=%s",
					varName,
					value,
				)
			}

			if service.BuildArgs == nil {
				service.BuildArgs = make(map[string]*string)
			}

			service.BuildArgs[varName] = &replacedValue
		}
	}

	return nil
}

func (c *gitCompose) fixOutput(ctx context.Context, objs *[]runtime.Object, namespace string) ([]runtime.Object, error) {
	extraObjs := []runtime.Object{}

	for _, obj := range *objs {
		c.fixNamespace(obj, namespace)

		deploymentExtraObjs, err := c.fixDeployment(ctx, obj)
		if err != nil {
			return nil, errors.Wrap(err, "fail to fix deployment")
		}
		extraObjs = append(extraObjs, deploymentExtraObjs...)

		secretObj, err := c.getSecretForImage(ctx, obj, namespace)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to get secret for image")
		}
		if secretObj != nil {
			extraObjs = append(extraObjs, secretObj)
		}
	}

	return extraObjs, nil
}

func (c *gitCompose) fixDeployment(ctx context.Context, obj runtime.Object) ([]runtime.Object, error) {
	deployment, ok := obj.(*appsv1.Deployment)
	if !ok {
		return nil, nil
	}

	extraObjs := []runtime.Object{}

	c.addLabels(deployment)
	c.addSecurityRestrictions(deployment)
	c.addNodeContraints(deployment)
	c.fixRestartPolicy(deployment)
	c.fixPullPolicy(deployment)
	c.addResourceLimits(deployment)
	c.removeHostPort(deployment)

	envVarsSecret, err := c.addEnvVars(ctx, deployment)
	if err != nil {
		return nil, errors.Wrap(err, "fail to add env vars")
	}
	if envVarsSecret != nil {
		extraObjs = append(extraObjs, envVarsSecret)
	}

	return extraObjs, nil
}

func (c *gitCompose) fixNamespace(obj runtime.Object, namespace string) {
	objMeta := obj.(metav1.Object)
	objMeta.SetNamespace(namespace)
}

func (c *gitCompose) addSecurityRestrictions(deployment *appsv1.Deployment) {
	podSpec := &deployment.Spec.Template.Spec
	podSpec.AutomountServiceAccountToken = pointer.Bool(false)

	podSpec.SecurityContext = &corev1.PodSecurityContext{
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}
}

func (c *gitCompose) addNodeContraints(deployment *appsv1.Deployment) {
	podSpec := &deployment.Spec.Template.Spec

	if podSpec.NodeSelector == nil {
		podSpec.NodeSelector = map[string]string{}
	}
	podSpec.NodeSelector["preview.ergomake.dev/role"] = "preview"

	podSpec.Tolerations = append(podSpec.Tolerations, corev1.Toleration{
		Key:      "preview.ergomake.dev/domain",
		Operator: "Equal",
		Value:    "previews",
		Effect:   "NoSchedule",
	})
}

func (c *gitCompose) fixRestartPolicy(deployment *appsv1.Deployment) {
	podSpec := &deployment.Spec.Template.Spec
	if podSpec.RestartPolicy != "" {
		// since this is a deployment, the only supported values are "Always" or "Never"
		// for us it makes sense to leave it as "Always"
		podSpec.RestartPolicy = "Always"
	}
}

func (c *gitCompose) fixPullPolicy(deployment *appsv1.Deployment) {
	podSpec := &deployment.Spec.Template.Spec
	for i := range podSpec.Containers {
		podSpec.Containers[i].ImagePullPolicy = "IfNotPresent"
	}
}

func (c *gitCompose) addResourceLimits(deployment *appsv1.Deployment) {
	podSpec := &deployment.Spec.Template.Spec
	for i := range podSpec.Containers {
		podSpec.Containers[i].Resources.Limits = corev1.ResourceList{
			corev1.ResourceEphemeralStorage: resource.MustParse("5Gi"),
			corev1.ResourceMemory:           resource.MustParse("1Gi"),
		}

		podSpec.Containers[i].Resources.Requests = corev1.ResourceList{
			corev1.ResourceEphemeralStorage: resource.MustParse("5Gi"),
			corev1.ResourceMemory:           resource.MustParse("1Gi"),
		}
	}
}

func (c *gitCompose) removeHostPort(deployment *appsv1.Deployment) {
	podSpec := &deployment.Spec.Template.Spec
	for i := range podSpec.Containers {
		for j := range podSpec.Containers[i].Ports {
			podSpec.Containers[i].Ports[j].HostPort = 0
		}
	}
}

func (c *gitCompose) addLabels(deployment *appsv1.Deployment) {
	serviceName := deployment.GetLabels()["io.kompose.service"]
	service := c.environment.Services[serviceName]
	repo, _ := builder.ComputeRepoAndBuildPath(c.projectPath, c.configFilePath, service.Build, c.gitOptions.Repo)

	labels := map[string]string{
		"app":                          service.ID,
		"preview.ergomake.dev/id":      service.ID,
		"preview.ergomake.dev/service": serviceName,
		"preview.ergomake.dev/owner":   c.gitOptions.Owner,
		"preview.ergomake.dev/repo":    repo,
		"preview.ergomake.dev/sha":     c.gitOptions.SHA,
	}

	mergedDeploymentLabels := deployment.GetObjectMeta().GetLabels()
	for k, v := range labels {
		mergedDeploymentLabels[k] = v
	}
	deployment.SetLabels(mergedDeploymentLabels)

	mergedDeploymentAnnotations := deployment.GetObjectMeta().GetAnnotations()
	for k, v := range labels {
		mergedDeploymentAnnotations[k] = v
	}
	deployment.SetAnnotations(mergedDeploymentAnnotations)

	mergedPodLabels := deployment.Spec.Template.GetObjectMeta().GetLabels()
	for k, v := range labels {
		mergedPodLabels[k] = v
	}
	deployment.Spec.Template.SetLabels(mergedPodLabels)

	mergedPodAnnotations := deployment.Spec.Template.GetObjectMeta().GetAnnotations()
	for k, v := range labels {
		mergedPodAnnotations[k] = v
	}
	deployment.Spec.Template.SetAnnotations(mergedPodAnnotations)
}

func (c *gitCompose) addEnvVars(ctx context.Context, deployment *appsv1.Deployment) (*corev1.Secret, error) {
	service := c.environment.Services[deployment.GetLabels()["io.kompose.service"]]
	repo, _ := builder.ComputeRepoAndBuildPath(c.projectPath, c.configFilePath, service.Build, c.gitOptions.Repo)

	vars, err := c.envVarsProvider.ListByRepoBranch(ctx, c.gitOptions.Owner, repo, c.gitOptions.Branch)
	if err != nil {
		return nil, errors.Wrap(err, "fail to list env vars by repo")
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-env-vars-secret", service.ID),
			Namespace: deployment.GetNamespace(),
		},
		Data: map[string][]byte{},
	}

	envVars := []corev1.EnvVar{}
	for _, v := range vars {
		secret.Data[v.Name] = []byte(v.Value)

		envVars = append(envVars, corev1.EnvVar{
			Name: v.Name,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secret.GetObjectMeta().GetName(),
					},
					Key: v.Name,
				},
			},
		})
	}

	podSpec := &deployment.Spec.Template.Spec
	for i := range podSpec.Containers {
		podSpec.Containers[i].Env = append(podSpec.Containers[i].Env, envVars...)
	}

	return secret, nil
}

func (c *gitCompose) getSecretForImage(
	ctx context.Context,
	obj runtime.Object,
	namespace string,
) (runtime.Object, error) {
	deployment, ok := obj.(*appsv1.Deployment)
	if !ok {
		return nil, nil
	}

	image := deployment.Spec.Template.Spec.Containers[0].Image
	if strings.HasPrefix(image, userlandRegistry) {
		return nil, nil
	}

	creds, err := c.privRegistryProvider.FetchCreds(ctx, c.gitOptions.Owner, image)
	if errors.Is(err, privregistry.ErrRegistryNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrapf(err, "fail to fetch token for image %s", image)
	}

	token := base64.StdEncoding.EncodeToString([]byte(creds.Token))
	authJSON := []byte(fmt.Sprintf(`{"auths": {"%s": { "auth": "%s" }}}`, creds.URL, token))

	data := make(map[string][]byte)
	data[corev1.DockerConfigJsonKey] = authJSON

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.GetName() + "-dockerconfig",
			Namespace: deployment.GetNamespace(),
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: data,
	}

	deployment.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
		{
			Name: secret.GetName(),
		},
	}

	runtimeSecret := runtime.Object(secret)

	return runtimeSecret, nil
}

func (c *gitCompose) getLabels(serviceID, serviceName string) map[string]string {
	return map[string]string{
		"app":                              serviceID,
		"preview.ergomake.dev/id":          serviceID,
		"preview.ergomake.dev/service":     serviceName,
		"preview.ergomake.dev/owner":       c.gitOptions.Owner,
		"preview.ergomake.dev/branchOwner": c.gitOptions.BranchOwner,
		"preview.ergomake.dev/repo":        c.gitOptions.Repo,
		"preview.ergomake.dev/sha":         c.gitOptions.SHA,
	}
}
