package cluster

import (
	"context"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ClusterEnv struct {
	Namespace string
	Objects   []runtime.Object
}

func Deploy(ctx context.Context, client Client, env *ClusterEnv) (err error) {
	err = client.CreateNamespace(ctx, env.Namespace)

	if err != nil {
		return errors.Wrapf(err, "fail to create namespace %s", env.Namespace)
	}

	// this works like a rollback, it deletes the namespace when error
	defer func() {
		if err == nil {
			return
		}

		// TODO: what if we fail to delete the namespace?
		client.DeleteNamespace(ctx, env.Namespace)
	}()

	for _, obj := range env.Objects {
		switch obj := obj.(type) {

		case *corev1.Secret:
			err = client.CreateSecret(ctx, obj)
			if err != nil {
				return errors.Wrapf(err, "fail to create %s secret", obj.Name)
			}
		case *corev1.Service:
			err = client.CreateService(ctx, obj)
			if err != nil {
				return errors.Wrapf(err, "fail to create %s service", obj.Name)
			}
		case *appsv1.Deployment:
			err = client.CreateDeployment(ctx, obj)
			if err != nil {
				return errors.Wrapf(err, "fail to create %s deployment", obj.Name)
			}
		case *corev1.ConfigMap:
			err = client.CreateConfigMap(ctx, obj)
			if err != nil {
				return errors.Wrapf(err, "fail to create %s configmap", obj.Name)
			}
		case *networkingv1.Ingress:
			err = client.CreateIngress(ctx, obj)
			if err != nil {
				return errors.Wrapf(err, "fail to create %s ingress", obj.Name)
			}

		case *networkingv1.NetworkPolicy:
			// TODO: this is being ignored because the default policy disallows ingress from outside of the cluster
			// but, we probably should have a network policy
			//
			// _, err := clientset.NetworkingV1().NetworkPolicies(obj.GetNamespace()).Create(ctx, obj, metav1.CreateOptions{})
			// if err != nil {
			// 	return err
			// }
			continue
		// Add cases for other object types here
		default:
			return errors.Errorf("unknown object type: %T", obj)
		}
	}

	return nil
}
