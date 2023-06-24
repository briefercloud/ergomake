package cluster_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/ergomake/ergomake/internal/cluster"
	mocks "github.com/ergomake/ergomake/mocks/cluster"
)

type myOwnObject struct{}

func (o *myOwnObject) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}
func (o *myOwnObject) DeepCopyObject() runtime.Object {
	return o
}

func TestDeploy(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		client func(t *testing.T, env *cluster.ClusterEnv) cluster.Client
		env    *cluster.ClusterEnv
		errors bool
	}

	tt := []testCase{
		{
			name: "errors when namespace errors",
			client: func(t *testing.T, _ *cluster.ClusterEnv) cluster.Client {
				client := mocks.NewClient(t)
				client.EXPECT().CreateNamespace(mock.Anything, "namespace").Return(errors.New("rip"))
				return client
			},
			env: &cluster.ClusterEnv{
				Namespace: "namespace",
			},
			errors: true,
		},
		{
			name: "errors on unexpected object and deletes namespace",
			client: func(t *testing.T, _ *cluster.ClusterEnv) cluster.Client {
				client := mocks.NewClient(t)
				client.EXPECT().CreateNamespace(mock.Anything, "namespace").Return(nil)
				client.EXPECT().DeleteNamespace(mock.Anything, "namespace").Return(nil)
				return client
			},
			env: &cluster.ClusterEnv{
				Namespace: "namespace",
				Objects:   []runtime.Object{&myOwnObject{}},
			},
			errors: true,
		},
		{
			name: "creates all sorts of objects",
			client: func(t *testing.T, env *cluster.ClusterEnv) cluster.Client {
				client := mocks.NewClient(t)
				client.EXPECT().CreateNamespace(mock.Anything, "namespace").Return(nil)
				client.EXPECT().CreateService(mock.Anything, env.Objects[0]).Return(nil)
				client.EXPECT().CreateDeployment(mock.Anything, env.Objects[1]).Return(nil)
				client.EXPECT().CreateConfigMap(mock.Anything, env.Objects[2]).Return(nil)
				client.EXPECT().CreateIngress(mock.Anything, env.Objects[3]).Return(nil)
				client.EXPECT().CreateIngress(mock.Anything, env.Objects[3]).Return(nil)
				client.EXPECT().CreateSecret(mock.Anything, env.Objects[4]).Return(nil)
				return client
			},
			env: &cluster.ClusterEnv{
				Namespace: "namespace",
				Objects: []runtime.Object{
					&corev1.Service{},
					&appsv1.Deployment{},
					&corev1.ConfigMap{},
					&networkingv1.Ingress{},
					&corev1.Secret{},
					&networkingv1.NetworkPolicy{},
				},
			},
			errors: false,
		},
	}

	failures := []struct {
		obj    runtime.Object
		method string
	}{
		{&corev1.Service{}, "CreateService"},
		{&appsv1.Deployment{}, "CreateDeployment"},
		{&corev1.ConfigMap{}, "CreateConfigMap"},
		{&networkingv1.Ingress{}, "CreateIngress"},
		{&corev1.Secret{}, "CreateSecret"},
	}
	for _, failure := range failures {
		func(method string, obj runtime.Object) {
			tt = append(tt, testCase{
				name: fmt.Sprintf("errors when %s errors and deletes namespace", failure.method),
				client: func(t *testing.T, env *cluster.ClusterEnv) cluster.Client {
					client := mocks.NewClient(t)

					client.EXPECT().CreateNamespace(mock.Anything, "namespace").Return(nil)
					client.EXPECT().DeleteNamespace(mock.Anything, "namespace").Return(nil)
					client.On(method, mock.Anything, env.Objects[0]).Return(errors.New("rip"))

					return client
				},
				env: &cluster.ClusterEnv{
					Namespace: "namespace",
					Objects:   []runtime.Object{obj},
				},
				errors: true,
			})
		}(failure.method, failure.obj)
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := cluster.Deploy(context.TODO(), tc.client(t, tc.env), tc.env)
			if tc.errors {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
