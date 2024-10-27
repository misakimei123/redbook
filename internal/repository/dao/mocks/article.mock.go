// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/repository/dao/article.go
//
// Generated by this command:
//
//	mockgen -source=./internal/repository/dao/article.go -package=daomocks -destination=./internal/repository/dao/mocks/article.mock.go
//

// Package daomocks is a generated GoMock package.
package daomocks

import (
	context "context"
	reflect "reflect"

	dao "github.com/misakimei123/redbook/internal/repository/dao"
	gomock "go.uber.org/mock/gomock"
)

// MockArticleDao is a mock of ArticleDao interface.
type MockArticleDao struct {
	ctrl     *gomock.Controller
	recorder *MockArticleDaoMockRecorder
}

// MockArticleDaoMockRecorder is the mock recorder for MockArticleDao.
type MockArticleDaoMockRecorder struct {
	mock *MockArticleDao
}

// NewMockArticleDao creates a new mock instance.
func NewMockArticleDao(ctrl *gomock.Controller) *MockArticleDao {
	mock := &MockArticleDao{ctrl: ctrl}
	mock.recorder = &MockArticleDaoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockArticleDao) EXPECT() *MockArticleDaoMockRecorder {
	return m.recorder
}

// GetByAuthor mocks base method.
func (m *MockArticleDao) GetByAuthor(ctx context.Context, uid int64, offset, limit int) ([]dao.Article, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByAuthor", ctx, uid, offset, limit)
	ret0, _ := ret[0].([]dao.Article)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByAuthor indicates an expected call of GetByAuthor.
func (mr *MockArticleDaoMockRecorder) GetByAuthor(ctx, uid, offset, limit any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByAuthor", reflect.TypeOf((*MockArticleDao)(nil).GetByAuthor), ctx, uid, offset, limit)
}

// GetById mocks base method.
func (m *MockArticleDao) GetById(ctx context.Context, uid, id int64) (dao.Article, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetById", ctx, uid, id)
	ret0, _ := ret[0].(dao.Article)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetById indicates an expected call of GetById.
func (mr *MockArticleDaoMockRecorder) GetById(ctx, uid, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetById", reflect.TypeOf((*MockArticleDao)(nil).GetById), ctx, uid, id)
}

// GetPubById mocks base method.
func (m *MockArticleDao) GetPubById(ctx context.Context, id int64) (dao.PublishedArticle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPubById", ctx, id)
	ret0, _ := ret[0].(dao.PublishedArticle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPubById indicates an expected call of GetPubById.
func (mr *MockArticleDaoMockRecorder) GetPubById(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPubById", reflect.TypeOf((*MockArticleDao)(nil).GetPubById), ctx, id)
}

// GetPubs mocks base method.
func (m *MockArticleDao) GetPubs(ctx context.Context, start, end int64, offset, limit int) ([]dao.PublishedArticle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPubs", ctx, start, end, offset, limit)
	ret0, _ := ret[0].([]dao.PublishedArticle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPubs indicates an expected call of GetPubs.
func (mr *MockArticleDaoMockRecorder) GetPubs(ctx, start, end, offset, limit any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPubs", reflect.TypeOf((*MockArticleDao)(nil).GetPubs), ctx, start, end, offset, limit)
}

// Insert mocks base method.
func (m *MockArticleDao) Insert(ctx context.Context, article dao.Article) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", ctx, article)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Insert indicates an expected call of Insert.
func (mr *MockArticleDaoMockRecorder) Insert(ctx, article any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*MockArticleDao)(nil).Insert), ctx, article)
}

// Sync mocks base method.
func (m *MockArticleDao) Sync(ctx context.Context, art dao.Article) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sync", ctx, art)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Sync indicates an expected call of Sync.
func (mr *MockArticleDaoMockRecorder) Sync(ctx, art any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sync", reflect.TypeOf((*MockArticleDao)(nil).Sync), ctx, art)
}

// SyncStatus mocks base method.
func (m *MockArticleDao) SyncStatus(ctx context.Context, id, uid int64, status uint8) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncStatus", ctx, id, uid, status)
	ret0, _ := ret[0].(error)
	return ret0
}

// SyncStatus indicates an expected call of SyncStatus.
func (mr *MockArticleDaoMockRecorder) SyncStatus(ctx, id, uid, status any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncStatus", reflect.TypeOf((*MockArticleDao)(nil).SyncStatus), ctx, id, uid, status)
}

// SyncV1 mocks base method.
func (m *MockArticleDao) SyncV1(ctx context.Context, art dao.Article) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SyncV1", ctx, art)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SyncV1 indicates an expected call of SyncV1.
func (mr *MockArticleDaoMockRecorder) SyncV1(ctx, art any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SyncV1", reflect.TypeOf((*MockArticleDao)(nil).SyncV1), ctx, art)
}

// UpdateById mocks base method.
func (m *MockArticleDao) UpdateById(ctx context.Context, article dao.Article) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateById", ctx, article)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateById indicates an expected call of UpdateById.
func (mr *MockArticleDaoMockRecorder) UpdateById(ctx, article any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateById", reflect.TypeOf((*MockArticleDao)(nil).UpdateById), ctx, article)
}