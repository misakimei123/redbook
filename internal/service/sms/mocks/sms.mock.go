// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/service/sms/type.go
//
// Generated by this command:
//
//	mockgen -source=./internal/service/sms/type.go -package=smsmocks -destination=./internal/service/sms/mocks/sms.mock.go
//

// Package smsmocks is a generated GoMock package.
package smsmocks

import (
	context "context"
	reflect "reflect"

	sms "github.com/misakimei123/redbook/internal/service/sms"
	gomock "go.uber.org/mock/gomock"
)

// MockSMSService is a mock of SMSService interface.
type MockSMSService struct {
	ctrl     *gomock.Controller
	recorder *MockSMSServiceMockRecorder
}

// MockSMSServiceMockRecorder is the mock recorder for MockSMSService.
type MockSMSServiceMockRecorder struct {
	mock *MockSMSService
}

// NewMockSMSService creates a new mock instance.
func NewMockSMSService(ctrl *gomock.Controller) *MockSMSService {
	mock := &MockSMSService{ctrl: ctrl}
	mock.recorder = &MockSMSServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSMSService) EXPECT() *MockSMSServiceMockRecorder {
	return m.recorder
}

// Send mocks base method.
func (m *MockSMSService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	m.ctrl.T.Helper()
	varargs := []any{ctx, tplId, args}
	for _, a := range number {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Send", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockSMSServiceMockRecorder) Send(ctx, tplId, args any, number ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, tplId, args}, number...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockSMSService)(nil).Send), varargs...)
}

// MockSMSBuilder is a mock of SMSBuilder interface.
type MockSMSBuilder struct {
	ctrl     *gomock.Controller
	recorder *MockSMSBuilderMockRecorder
}

// MockSMSBuilderMockRecorder is the mock recorder for MockSMSBuilder.
type MockSMSBuilderMockRecorder struct {
	mock *MockSMSBuilder
}

// NewMockSMSBuilder creates a new mock instance.
func NewMockSMSBuilder(ctrl *gomock.Controller) *MockSMSBuilder {
	mock := &MockSMSBuilder{ctrl: ctrl}
	mock.recorder = &MockSMSBuilderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSMSBuilder) EXPECT() *MockSMSBuilderMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockSMSBuilder) Create(model sms.SMSServiceType) sms.SMSService {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", model)
	ret0, _ := ret[0].(sms.SMSService)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockSMSBuilderMockRecorder) Create(model any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockSMSBuilder)(nil).Create), model)
}