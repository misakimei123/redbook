// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/service/sms/repo/sms.go
//
// Generated by this command:
//
//	mockgen -source=./internal/service/sms/repo/sms.go -package=smsrepomocks -destination=./internal/service/sms/repo/mock/sms.mock.go
//

// Package smsrepomocks is a generated GoMock package.
package smsrepomocks

import (
	context "context"
	reflect "reflect"

	repo "github.com/misakimei123/redbook/internal/service/sms/repo"
	gomock "go.uber.org/mock/gomock"
)

// MockSMSRepo is a mock of SMSRepo interface.
type MockSMSRepo struct {
	ctrl     *gomock.Controller
	recorder *MockSMSRepoMockRecorder
}

// MockSMSRepoMockRecorder is the mock recorder for MockSMSRepo.
type MockSMSRepoMockRecorder struct {
	mock *MockSMSRepo
}

// NewMockSMSRepo creates a new mock instance.
func NewMockSMSRepo(ctrl *gomock.Controller) *MockSMSRepo {
	mock := &MockSMSRepo{ctrl: ctrl}
	mock.recorder = &MockSMSRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSMSRepo) EXPECT() *MockSMSRepoMockRecorder {
	return m.recorder
}

// Del4Fail mocks base method.
func (m *MockSMSRepo) Del4Fail(ctx context.Context, id int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Del4Fail", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Del4Fail indicates an expected call of Del4Fail.
func (mr *MockSMSRepoMockRecorder) Del4Fail(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Del4Fail", reflect.TypeOf((*MockSMSRepo)(nil).Del4Fail), ctx, id)
}

// Del4Success mocks base method.
func (m *MockSMSRepo) Del4Success(ctx context.Context, id int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Del4Success", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Del4Success indicates an expected call of Del4Success.
func (mr *MockSMSRepoMockRecorder) Del4Success(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Del4Success", reflect.TypeOf((*MockSMSRepo)(nil).Del4Success), ctx, id)
}

// Get mocks base method.
func (m *MockSMSRepo) Get(ctx context.Context) (repo.SMSPara, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx)
	ret0, _ := ret[0].(repo.SMSPara)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockSMSRepoMockRecorder) Get(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockSMSRepo)(nil).Get), ctx)
}

// Put mocks base method.
func (m *MockSMSRepo) Put(ctx context.Context, para repo.SMSPara) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Put", ctx, para)
	ret0, _ := ret[0].(error)
	return ret0
}

// Put indicates an expected call of Put.
func (mr *MockSMSRepoMockRecorder) Put(ctx, para any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockSMSRepo)(nil).Put), ctx, para)
}