// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	json "encoding/json"

	mock "github.com/stretchr/testify/mock"
)

// ProjectValidationError is an autogenerated mock type for the ProjectValidationError type
type ProjectValidationError struct {
	mock.Mock
}

type ProjectValidationError_Expecter struct {
	mock *mock.Mock
}

func (_m *ProjectValidationError) EXPECT() *ProjectValidationError_Expecter {
	return &ProjectValidationError_Expecter{mock: &_m.Mock}
}

// GetErrorMessage provides a mock function with given fields:
func (_m *ProjectValidationError) GetErrorMessage() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ProjectValidationError_GetErrorMessage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetErrorMessage'
type ProjectValidationError_GetErrorMessage_Call struct {
	*mock.Call
}

// GetErrorMessage is a helper method to define mock.On call
func (_e *ProjectValidationError_Expecter) GetErrorMessage() *ProjectValidationError_GetErrorMessage_Call {
	return &ProjectValidationError_GetErrorMessage_Call{Call: _e.mock.On("GetErrorMessage")}
}

func (_c *ProjectValidationError_GetErrorMessage_Call) Run(run func()) *ProjectValidationError_GetErrorMessage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ProjectValidationError_GetErrorMessage_Call) Return(_a0 string) *ProjectValidationError_GetErrorMessage_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ProjectValidationError_GetErrorMessage_Call) RunAndReturn(run func() string) *ProjectValidationError_GetErrorMessage_Call {
	_c.Call.Return(run)
	return _c
}

// Serialize provides a mock function with given fields:
func (_m *ProjectValidationError) Serialize() json.RawMessage {
	ret := _m.Called()

	var r0 json.RawMessage
	if rf, ok := ret.Get(0).(func() json.RawMessage); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(json.RawMessage)
		}
	}

	return r0
}

// ProjectValidationError_Serialize_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Serialize'
type ProjectValidationError_Serialize_Call struct {
	*mock.Call
}

// Serialize is a helper method to define mock.On call
func (_e *ProjectValidationError_Expecter) Serialize() *ProjectValidationError_Serialize_Call {
	return &ProjectValidationError_Serialize_Call{Call: _e.mock.On("Serialize")}
}

func (_c *ProjectValidationError_Serialize_Call) Run(run func()) *ProjectValidationError_Serialize_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ProjectValidationError_Serialize_Call) Return(_a0 json.RawMessage) *ProjectValidationError_Serialize_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ProjectValidationError_Serialize_Call) RunAndReturn(run func() json.RawMessage) *ProjectValidationError_Serialize_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewProjectValidationError interface {
	mock.TestingT
	Cleanup(func())
}

// NewProjectValidationError creates a new instance of ProjectValidationError. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewProjectValidationError(t mockConstructorTestingTNewProjectValidationError) *ProjectValidationError {
	mock := &ProjectValidationError{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}