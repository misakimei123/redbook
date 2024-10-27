// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/repository/article_reader.go
//
// Generated by this command:
//
//	mockgen -source=./internal/repository/article_reader.go -package=repomocks -destination=./internal/repository/mocks/article_reader.mock.go
//

// Package repomocks is a generated GoMock package.
package repomocks

import (
	context "context"
	reflect "reflect"

	domain "github.com/misakimei123/redbook/internal/domain"
	gomock "go.uber.org/mock/gomock"
)

// MockArticleReaderRepository is a mock of ArticleReaderRepository interface.
type MockArticleReaderRepository struct {
	ctrl     *gomock.Controller
	recorder *MockArticleReaderRepositoryMockRecorder
}

// MockArticleReaderRepositoryMockRecorder is the mock recorder for MockArticleReaderRepository.
type MockArticleReaderRepositoryMockRecorder struct {
	mock *MockArticleReaderRepository
}

// NewMockArticleReaderRepository creates a new mock instance.
func NewMockArticleReaderRepository(ctrl *gomock.Controller) *MockArticleReaderRepository {
	mock := &MockArticleReaderRepository{ctrl: ctrl}
	mock.recorder = &MockArticleReaderRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockArticleReaderRepository) EXPECT() *MockArticleReaderRepositoryMockRecorder {
	return m.recorder
}

// Save mocks base method.
func (m *MockArticleReaderRepository) Save(ctx context.Context, article domain.Article) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", ctx, article)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockArticleReaderRepositoryMockRecorder) Save(ctx, article any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockArticleReaderRepository)(nil).Save), ctx, article)
}

// Update mocks base method.
func (m *MockArticleReaderRepository) Update(ctx context.Context, article domain.Article) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, article)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockArticleReaderRepositoryMockRecorder) Update(ctx, article any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockArticleReaderRepository)(nil).Update), ctx, article)
}
