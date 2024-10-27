package service

import (
	"context"
	"errors"
	"testing"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/repository"
	repomocks "github.com/misakimei123/redbook/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestArticleService_Publish(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(controller *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository)
		art     domain.Article
		wantId  int64
		wantErr error
	}{
		{
			name: "publish success",
			mock: func(controller *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				articleAuthorRepository := repomocks.NewMockArticleAuthorRepository(controller)
				art := domain.Article{
					Title:   "my title",
					Content: "my content",
					Author:  domain.Author{Id: 123},
					Status:  domain.ArticleStatusPublished,
				}
				articleAuthorRepository.EXPECT().Create(gomock.Any(), art).Return(int64(1), nil)
				art.Id = 1
				articleReaderRepository := repomocks.NewMockArticleReaderRepository(controller)
				articleReaderRepository.EXPECT().Save(gomock.Any(), art).Return(nil)
				return articleAuthorRepository, articleReaderRepository
			},
			art: domain.Article{
				Title:   "my title",
				Content: "my content",
				Author:  domain.Author{Id: 123},
				Status:  domain.ArticleStatusPublished,
			},
			wantId:  1,
			wantErr: nil,
		},
		{
			name: "modify and publish success",
			mock: func(controller *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				articleAuthorRepository := repomocks.NewMockArticleAuthorRepository(controller)
				art := domain.Article{
					Id:      1,
					Title:   "my title",
					Content: "my content",
					Author:  domain.Author{Id: 123},
					Status:  domain.ArticleStatusPublished,
				}
				articleAuthorRepository.EXPECT().Update(gomock.Any(), art).Return(nil)
				articleReaderRepository := repomocks.NewMockArticleReaderRepository(controller)
				articleReaderRepository.EXPECT().Save(gomock.Any(), art).Return(nil)
				return articleAuthorRepository, articleReaderRepository
			},
			art: domain.Article{
				Id:      1,
				Title:   "my title",
				Content: "my content",
				Author:  domain.Author{Id: 123},
			},
			wantId:  1,
			wantErr: nil,
		},
		{
			name: "modify and publish fail, try again success",
			mock: func(controller *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				articleAuthorRepository := repomocks.NewMockArticleAuthorRepository(controller)
				art := domain.Article{
					Id:      1,
					Title:   "my title",
					Content: "my content",
					Author:  domain.Author{Id: 123},
					Status:  domain.ArticleStatusPublished,
				}
				articleAuthorRepository.EXPECT().Update(gomock.Any(), art).Return(nil)
				articleReaderRepository := repomocks.NewMockArticleReaderRepository(controller)
				articleReaderRepository.EXPECT().Save(gomock.Any(), art).Return(errors.New("mock db fail"))
				articleReaderRepository.EXPECT().Save(gomock.Any(), art).Return(nil)
				return articleAuthorRepository, articleReaderRepository
			},
			art: domain.Article{
				Id:      1,
				Title:   "my title",
				Content: "my content",
				Author:  domain.Author{Id: 123},
			},
			wantId:  1,
			wantErr: nil,
		},
		{
			name: "modify and publish fail",
			mock: func(controller *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				articleAuthorRepository := repomocks.NewMockArticleAuthorRepository(controller)
				art := domain.Article{
					Id:      1,
					Title:   "my title",
					Content: "my content",
					Author:  domain.Author{Id: 123},
					Status:  domain.ArticleStatusPublished,
				}
				articleAuthorRepository.EXPECT().Update(gomock.Any(), art).Return(nil)
				articleReaderRepository := repomocks.NewMockArticleReaderRepository(controller)
				articleReaderRepository.EXPECT().Save(gomock.Any(), art).Times(3).Return(errors.New("mock db fail"))
				return articleAuthorRepository, articleReaderRepository
			},
			art: domain.Article{
				Id:      1,
				Title:   "my title",
				Content: "my content",
				Author:  domain.Author{Id: 123},
			},
			wantId:  1,
			wantErr: errors.New("多次保存线上库失败"),
		},
		{
			name: "save author fail",
			mock: func(controller *gomock.Controller) (repository.ArticleAuthorRepository, repository.ArticleReaderRepository) {
				articleAuthorRepository := repomocks.NewMockArticleAuthorRepository(controller)
				art := domain.Article{
					Id:      1,
					Title:   "my title",
					Content: "my content",
					Author:  domain.Author{Id: 123},
					Status:  domain.ArticleStatusPublished,
				}
				articleAuthorRepository.EXPECT().Update(gomock.Any(), art).Return(errors.New("mock db fail"))
				articleReaderRepository := repomocks.NewMockArticleReaderRepository(controller)
				return articleAuthorRepository, articleReaderRepository
			},
			art: domain.Article{
				Id:      1,
				Title:   "my title",
				Content: "my content",
				Author:  domain.Author{Id: 123},
			},
			wantId:  0,
			wantErr: errors.New("mock db fail"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			authorRepository, readerRepository := tc.mock(ctrl)
			svc := NewArticleServiceV1(authorRepository, readerRepository)
			id, err := svc.PublishV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
