// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	context "context"

	cluster "github.com/ergomake/ergomake/internal/cluster"

	mock "github.com/stretchr/testify/mock"

	transformer "github.com/ergomake/ergomake/internal/transformer"
)

// Transformer is an autogenerated mock type for the Transformer type
type Transformer struct {
	mock.Mock
}

type Transformer_Expecter struct {
	mock *mock.Mock
}

func (_m *Transformer) EXPECT() *Transformer_Expecter {
	return &Transformer_Expecter{mock: &_m.Mock}
}

// Transform provides a mock function with given fields: ctx, namespace
func (_m *Transformer) Transform(ctx context.Context, namespace string) (*cluster.ClusterEnv, *transformer.Environment, error) {
	ret := _m.Called(ctx, namespace)

	var r0 *cluster.ClusterEnv
	var r1 *transformer.Environment
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*cluster.ClusterEnv, *transformer.Environment, error)); ok {
		return rf(ctx, namespace)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *cluster.ClusterEnv); ok {
		r0 = rf(ctx, namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*cluster.ClusterEnv)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) *transformer.Environment); ok {
		r1 = rf(ctx, namespace)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*transformer.Environment)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, string) error); ok {
		r2 = rf(ctx, namespace)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Transformer_Transform_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Transform'
type Transformer_Transform_Call struct {
	*mock.Call
}

// Transform is a helper method to define mock.On call
//   - ctx context.Context
//   - namespace string
func (_e *Transformer_Expecter) Transform(ctx interface{}, namespace interface{}) *Transformer_Transform_Call {
	return &Transformer_Transform_Call{Call: _e.mock.On("Transform", ctx, namespace)}
}

func (_c *Transformer_Transform_Call) Run(run func(ctx context.Context, namespace string)) *Transformer_Transform_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Transformer_Transform_Call) Return(_a0 *cluster.ClusterEnv, _a1 *transformer.Environment, _a2 error) *Transformer_Transform_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *Transformer_Transform_Call) RunAndReturn(run func(context.Context, string) (*cluster.ClusterEnv, *transformer.Environment, error)) *Transformer_Transform_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewTransformer interface {
	mock.TestingT
	Cleanup(func())
}

// NewTransformer creates a new instance of Transformer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTransformer(t mockConstructorTestingTNewTransformer) *Transformer {
	mock := &Transformer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
