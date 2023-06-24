package cluster

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/utils/pointer"
)

type k8sClient struct {
	*kubernetes.Clientset
}

func k8sConfig() (*rest.Config, error) {
	configPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		// give empty to try in cluster config
		configPath = ""
	}

	return clientcmd.BuildConfigFromFlags("", configPath)
}

func NewK8sClient() (*k8sClient, error) {
	config, err := k8sConfig()
	if err != nil {
		return nil, errors.Wrap(err, "fail to build k8s config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create k8s clientset")
	}

	return &k8sClient{clientset}, nil
}

func (k8s *k8sClient) CreateNamespace(ctx context.Context, namespace string) error {
	namespaceObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"dev.ergomake.env/type": "preview",
			},
		},
	}
	_, err := k8s.CoreV1().Namespaces().
		Create(ctx, namespaceObj, metav1.CreateOptions{})

	return err
}

func (k8s *k8sClient) GetPreviewNamespaces(ctx context.Context) ([]corev1.Namespace, error) {
	namespaces, err := k8s.CoreV1().Namespaces().
		List(ctx, metav1.ListOptions{
			LabelSelector: "dev.ergomake.env/type=preview",
		})
	if err != nil {
		return nil, errors.Wrap(err, "fail to list namespaces")
	}

	return namespaces.Items, nil
}

func (k8s *k8sClient) DeleteNamespace(ctx context.Context, namespace string) error {
	return k8s.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
}

func (k8s *k8sClient) CreateService(ctx context.Context, service *corev1.Service) error {
	_, err := k8s.CoreV1().Services(service.GetNamespace()).
		Create(ctx, service, metav1.CreateOptions{})

	return err
}

func (k8s *k8sClient) CreateDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	_, err := k8s.AppsV1().Deployments(deployment.GetNamespace()).
		Create(ctx, deployment, metav1.CreateOptions{})

	return err
}

func (k8s *k8sClient) CreateConfigMap(ctx context.Context, configMap *corev1.ConfigMap) error {
	_, err := k8s.CoreV1().ConfigMaps(configMap.GetNamespace()).
		Create(ctx, configMap, metav1.CreateOptions{})

	return err
}

func (k8s *k8sClient) CreateIngress(ctx context.Context, ingress *networkingv1.Ingress) error {
	_, err := k8s.NetworkingV1().Ingresses(ingress.GetNamespace()).
		Create(ctx, ingress, metav1.CreateOptions{})

	return err
}

func (k8s *k8sClient) CreateJob(ctx context.Context, job *batchv1.Job) (*batchv1.Job, error) {
	return k8s.BatchV1().Jobs(job.GetNamespace()).Create(ctx, job, metav1.CreateOptions{})
}

func (k8s *k8sClient) CreateSecret(ctx context.Context, secret *corev1.Secret) error {
	_, err := k8s.CoreV1().Secrets(secret.GetNamespace()).
		Create(ctx, secret, metav1.CreateOptions{})

	return err
}

func (k8s *k8sClient) GetIngressUrl(ctx context.Context, namespace string, serviceName string, protocol string) (string, error) {
	ingresses, err := k8s.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, ingress := range ingresses.Items {
		if ingress.Name == serviceName {
			if len(ingress.Spec.Rules) > 0 && len(ingress.Spec.Rules[0].Host) > 0 {
				return protocol + "://" + ingress.Spec.Rules[0].Host, nil
			}
			break
		}
	}

	return "", ErrIngressNotFound
}

func (k8s *k8sClient) GetIngress(ctx context.Context, namespace, name string) (*networkingv1.Ingress, error) {
	ingress, err := k8s.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return nil, ErrIngressNotFound
	}

	return ingress, err
}

func (k8s *k8sClient) GetDeployment(ctx context.Context, namespace string, deploymentName string) (*appsv1.Deployment, error) {
	return k8s.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
}

// WaitJobs waits for the specified jobs to complete. If the context passed as an argument
// contains a deadline, the function will time out and return an error if the deadline is exceeded.
// If the context does not have a deadline, the function will use a 2-minute timeout. The function
// returns a WaitJobsResult struct with the status of the completed jobs, or an error if the
// function times out or encounters an error.
func (k8s *k8sClient) WaitJobs(ctx context.Context, jobs []*batchv1.Job) (*WaitJobsResult, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
	}

	completed := make(map[string]struct{})
	result := &WaitJobsResult{}

	for {
		for _, job := range jobs {
			if _, ok := completed[job.Name]; ok {
				continue
			}

			job, err := k8s.BatchV1().Jobs(job.GetNamespace()).
				Get(ctx, job.GetName(), metav1.GetOptions{})
			if err != nil {
				return result, errors.Wrapf(
					err,
					"failed to get job %s status",
					job.GetName(),
				)
			}

			if job.Status.Succeeded > 0 || job.Status.Failed > 0 {
				completed[job.Name] = struct{}{}
			}

			if job.Status.Failed > 0 {
				result.Failed = append(result.Failed, job)
			}

			if job.Status.Succeeded > 0 {
				result.Succeeded = append(result.Succeeded, job)
			}
		}

		if len(completed) == len(jobs) {
			return result, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// continue with the next iteration of the loop
		}

		time.Sleep(5 * time.Second)
	}
}

// WaitDeployments waits for all deployments in the specified namespace to be ready.
// If the context passed as an argument contains a deadline, the function will use it as the timeout
// for waiting for all deployments to be ready. Otherwise, it will use a 2-minute timeout.
// The function checks the status of each deployment and returns an error if any of them have not
// yet reached the desired number of replicas. It returns nil when all deployments are ready.
func (k8s *k8sClient) WaitDeployments(ctx context.Context, namespace string) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
	}

	for {
		deployments, err := k8s.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}

		allReady := true
		for _, deployment := range deployments.Items {
			if deployment.Status.ReadyReplicas != *deployment.Spec.Replicas {
				allReady = false
				break
			}
		}

		if allReady {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		time.Sleep(5 * time.Second)
	}
}

func (k8s *k8sClient) GetJobLogs(ctx context.Context, job *batchv1.Job, size int64) (string, error) {
	pods, err := k8s.CoreV1().Pods(job.Namespace).List(ctx, metav1.ListOptions{
		TypeMeta:      metav1.TypeMeta{},
		LabelSelector: labels.Set(job.Spec.Selector.MatchLabels).String(),
	})
	if err != nil {
		return "", errors.Wrapf(err, "fail to list pods of job %s/%s", job.Namespace, job.Name)
	}

	if len(pods.Items) == 0 {
		return "", errors.Errorf("no pods found for job %s/%s", job.Namespace, job.Name)
	}

	containerName := ""
	for _, container := range job.Spec.Template.Spec.Containers {
		if container.Name == "" {
			containerName = "default"
			break
		} else {
			containerName = container.Name
		}
	}

	// sort the pods by creation timestamp, with the latest one first
	sort.Slice(pods.Items, func(i, j int) bool {
		return pods.Items[j].CreationTimestamp.Before(&pods.Items[i].CreationTimestamp)
	})

	req := k8s.CoreV1().Pods(pods.Items[0].Namespace).GetLogs(pods.Items[0].Name, &corev1.PodLogOptions{
		Container: containerName,
		// since a line has at least 1 char, reading `size` lines will give us at least `size` amount of chars
		TailLines: &size,
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, stream)

	if err != nil {
		return "", errors.Wrap(err, "fail to copy bytes from stream to buffer")
	}

	logs := buf.String()
	if int64(len(logs)) > size {
		logs = logs[int64(len(logs))-size:]
	}

	return logs, nil
}

func (k8s *k8sClient) ListJobs(ctx context.Context, namespace string) ([]*batchv1.Job, error) {
	jobList, err := k8s.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var jobs []*batchv1.Job
	for _, job := range jobList.Items {
		jobCopy := job // Create a copy of the job object
		jobs = append(jobs, &jobCopy)
	}

	return jobs, nil
}

func (k8s *k8sClient) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
	deployment, err := k8s.GetDeployment(ctx, namespace, name)
	if err != nil {
		return errors.Wrapf(err, "fail to get deployment %s at namespace %s", name, namespace)
	}

	deployment.Spec.Replicas = &replicas

	_, err = k8s.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})

	return errors.Wrap(err, "fail to update deployment replicas")
}

func (k8s *k8sClient) UpdateIngress(ctx context.Context, ingress *networkingv1.Ingress) error {
	_, err := k8s.NetworkingV1().Ingresses(ingress.GetNamespace()).Update(ctx, ingress, metav1.UpdateOptions{})

	return err
}

func (k8s *k8sClient) AreServicesAlive(ctx context.Context, namespace string) (bool, error) {
	services, err := k8s.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, errors.Wrapf(err, "fail to list services of namespace %s", namespace)
	}

	allReady := true
	for _, svc := range services.Items {
		for _, port := range svc.Spec.Ports {
			address := svc.Spec.ClusterIP + ":" + strconv.Itoa(int(port.Port))
			conn, connErr := net.DialTimeout("tcp", address, 1*time.Second)
			if connErr != nil {
				allReady = false
				break
			}
			_ = conn.Close()
		}

		if !allReady {
			break
		}
	}

	return allReady, nil

}

func (k8s *k8sClient) WatchServiceLogs(ctx context.Context, namespace, name string, sinceSeconds int64) (<-chan string, <-chan error, error) {
	logsCh := make(chan string)
	errCh := make(chan error)

	service, err := k8s.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		close(logsCh)
		close(errCh)
		return logsCh, errCh, errors.Wrapf(err, "fail to get service %s/%s", namespace, name)
	}

	labelSelector := metav1.FormatLabelSelector(&metav1.LabelSelector{
		MatchLabels: service.Spec.Selector,
	})

	pods, err := k8s.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		close(logsCh)
		close(errCh)
		return logsCh, errCh, errors.Wrapf(err, "fail to list pods for service %s/%s", namespace, name)
	}

	watchContainerLogs := func(pod corev1.Pod, container corev1.Container) {
		podLogOpts := corev1.PodLogOptions{
			Container:    container.Name,
			Follow:       true,
			SinceSeconds: pointer.Int64(sinceSeconds),
		}

		req := k8s.CoreV1().Pods(namespace).GetLogs(pod.Name, &podLogOpts)
		stream, err := req.Stream(ctx)
		if err != nil {
			errCh <- err
			return
		}
		defer stream.Close()

		go func() {
			<-ctx.Done()
			stream.Close()
		}()

		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			line := scanner.Text()
			logsCh <- line
		}

		if err := scanner.Err(); err != nil {
			errCh <- err
		} else {
			errCh <- nil
		}
	}

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			go watchContainerLogs(pod, container)
		}
	}

	return logsCh, errCh, nil
}
