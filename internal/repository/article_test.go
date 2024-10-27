package repository

import (
	"context"
	"testing"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/repository/dao"
	daomocks "github.com/misakimei123/redbook/internal/repository/dao/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// repo层分发，不同库
func TestCachedArticleRepository_SyncV1(t *testing.T) {
	testCases := []struct {
		name    string
		art     domain.Article
		mock    func(controller *gomock.Controller) (dao.ArticleAuthorDao, dao.ArticleReaderDao)
		wantId  int64
		wantErr error
	}{
		{
			name: "new sync success",
			art: domain.Article{
				Title:   "my title",
				Content: "my content",
				Author:  domain.Author{Id: 123},
				Status:  domain.ArticleStatusPublished,
			},
			mock: func(ctrl *gomock.Controller) (dao.ArticleAuthorDao, dao.ArticleReaderDao) {
				art := dao.Article{
					Title:    "my title",
					Content:  "my content",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished,
				}
				authorDao := daomocks.NewMockArticleAuthorDao(ctrl)
				authorDao.EXPECT().Create(gomock.Any(), art).Return(int64(1), nil)
				readerDao := daomocks.NewMockArticleReaderDao(ctrl)
				art.Id = 1
				readerDao.EXPECT().Upsert(gomock.Any(), art).Return(nil)
				return authorDao, readerDao
			},
			wantId:  1,
			wantErr: nil,
		},
		{
			name: "modify sync success",
			art: domain.Article{
				Id:      2,
				Title:   "my title",
				Content: "my content",
				Author:  domain.Author{Id: 123},
				Status:  domain.ArticleStatusPublished,
			},
			mock: func(ctrl *gomock.Controller) (dao.ArticleAuthorDao, dao.ArticleReaderDao) {
				art := dao.Article{
					Id:       2,
					Title:    "my title",
					Content:  "my content",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished,
				}
				authorDao := daomocks.NewMockArticleAuthorDao(ctrl)
				authorDao.EXPECT().UpdateById(gomock.Any(), art).Return(nil)
				readerDao := daomocks.NewMockArticleReaderDao(ctrl)
				readerDao.EXPECT().Upsert(gomock.Any(), art).Return(nil)
				return authorDao, readerDao
			},
			wantId:  2,
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			authorDao, readerDao := tc.mock(ctrl)
			articleRepository := NewCachedArticleRepositoryV1(authorDao, readerDao)
			id, err := articleRepository.SyncV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
