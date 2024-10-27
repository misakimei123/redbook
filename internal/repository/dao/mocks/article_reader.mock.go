// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/repository/dao/article_reader.go
//
// Generated by this command:
//
//	mockgen -source=./internal/repository/dao/article_reader.go -package=daomocks -destination=./internal/repository/dao/mocks/article_reader.mock.go
//

// Package daomocks is a generated GoMock package.
package daomocks

import (
	context "context"
	reflect "reflect"

	dao "github.com/misakimei123/redbook/internal/repository/dao"
	gomock "go.uber.org/mock/gomock"
)

// MockArticleReaderDao is a mock of ArticleReaderDao interface.
type MockArticleReaderDao struct {
	ctrl     *gomock.Controller
	recorder *MockArticleReaderDaoMockRecorder
}

// MockArticleReaderDaoMockRecorder is the mock recorder for MockArticleReaderDao.
type MockArticleReaderDaoMockRecorder struct {
	mock *MockArticleReaderDao
}

// NewMockArticleReaderDao creates a new mock instance.
func NewMockArticleReaderDao(ctrl *gomock.Controller) *MockArticleReaderDao {
	mock := &MockArticleReaderDao{ctrl: ctrl}
	mock.recorder = &MockArticleReaderDaoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockArticleReaderDao) EXPECT() *MockArticleReaderDaoMockRecorder {
	return m.recorder
}

// Upsert mocks base method.
func (m *MockArticleReaderDao) Upsert(ctx context.Context, article dao.Article) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upsert", ctx, article)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upsert indicates an expected call of Upsert.
func (mr *MockArticleReaderDaoMockRecorder) Upsert(ctx, article any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upsert", reflect.TypeOf((*MockArticleReaderDao)(nil).Upsert), ctx, article)
}

// UpsertV2 mocks base method.
func (m *MockArticleReaderDao) UpsertV2(ctx context.Context, article dao.PublishedArticle) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertV2", ctx, article)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertV2 indicates an expected call of UpsertV2.
func (mr *MockArticleReaderDaoMockRecorder) UpsertV2(ctx, article any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertV2", reflect.TypeOf((*MockArticleReaderDao)(nil).UpsertV2), ctx, article)
}
