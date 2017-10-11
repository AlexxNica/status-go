// Code generated by MockGen. DO NOT EDIT.
// Source: geth/common/notification.go

// Package common is a generated GoMock package.
package common

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockNotifier is a mock of Notifier interface
type MockNotifier struct {
	ctrl     *gomock.Controller
	recorder *MockNotifierMockRecorder
}

// MockNotifierMockRecorder is the mock recorder for MockNotifier
type MockNotifierMockRecorder struct {
	mock *MockNotifier
}

// NewMockNotifier creates a new mock instance
func NewMockNotifier(ctrl *gomock.Controller) *MockNotifier {
	mock := &MockNotifier{ctrl: ctrl}
	mock.recorder = &MockNotifierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNotifier) EXPECT() *MockNotifierMockRecorder {
	return m.recorder
}

// Notify mocks base method
func (m *MockNotifier) Notify(body interface{}, tokens ...string) error {
	varargs := []interface{}{body}
	for _, a := range tokens {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Notify", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Notify indicates an expected call of Notify
func (mr *MockNotifierMockRecorder) Notify(body interface{}, tokens ...interface{}) *gomock.Call {
	varargs := append([]interface{}{body}, tokens...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Notify", reflect.TypeOf((*MockNotifier)(nil).Notify), varargs...)
}
