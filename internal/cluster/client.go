package cluster

import (
	"context"

	kpack "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

var ErrIngressNotFound = errors.New("ingress not found")

type WaitJobsResult struct {
	Failed    []*batchv1.Job
	Succeeded []*batchv1.Job
}

type Starter interface {
	Start(stopCh <-chan struct{})
}

type Client interface {
	CreateNamespace(ctx context.Context, namespace string) error
	DeleteNamespace(ctx context.Context, namespace string) error
	CreateService(ctx context.Context, service *corev1.Service) error
	CreateDeployment(ctx context.Context, deployment *appsv1.Deployment) error
	CreateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) error
	CreateIngress(ctx context.Context, ingress *networkingv1.Ingress) error
	CreateJob(ctx context.Context, job *batchv1.Job) (*batchv1.Job, error)
	CreateSecret(ctx context.Context, secret *corev1.Secret) error
	CreateServiceAccount(ctx context.Context, svcAcc *corev1.ServiceAccount) error
	GetPreviewNamespaces(ctx context.Context) ([]corev1.Namespace, error)
	GetIngress(ctx context.Context, namespace, name string) (*networkingv1.Ingress, error)
	GetIngressUrl(ctx context.Context, namespace string, serviceName string, protocol string) (string, error)
	UpdateIngress(ctx context.Context, ingress *networkingv1.Ingress) error
	GetDeployment(ctx context.Context, namespace string, deploymentName string) (*appsv1.Deployment, error)
	ScaleDeployment(ctx context.Context, namespace string, deploymentName string, replicas int32) error
	WaitJobs(ctx context.Context, jobs []*batchv1.Job) (*WaitJobsResult, error)
	WaitDeployments(ctx context.Context, namespace string) error
	GetJobLogs(ctx context.Context, job *batchv1.Job, size int64) (string, error)
	ListJobs(ctx context.Context, namespace string) ([]*batchv1.Job, error)
	AreServicesAlive(ctx context.Context, namespace string) (bool, error)
	WatchServiceLogs(ctx context.Context, namespace, name string, sinceSeconds int64) (<-chan string, <-chan error, error)
	ApplyKPackBuilds(ctx context.Context, builds []*kpack.Build) error
	WatchResource(ctx context.Context, gvr schema.GroupVersionResource, handler cache.ResourceEventHandlerFuncs) (Starter, error)
	CopySecret(ctx context.Context, fromNS, toNS, name string) (*corev1.Secret, error)
}
