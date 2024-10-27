package service

import (
	"context"
	"testing"
	"time"

	domain2 "github.com/misakimei123/redbook/interactive/domain"
	"github.com/misakimei123/redbook/interactive/service"
	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/repository"
	repomocks "github.com/misakimei123/redbook/internal/repository/mocks"
	svcmocks "github.com/misakimei123/redbook/internal/service/mocks"
	"github.com/misakimei123/redbook/pkg/logger"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestArticleRankingService_DoRanking(t *testing.T) {
	const batchSize = 2
	testCases := []struct {
		name     string
		mock     func(controller *gomock.Controller) (ArticleService, service.InteractiveService, repository.RankingRepository)
		wantArts []domain.Article
		wantErr  error
	}{
		{
			name: "success generate",
			mock: func(ctrl *gomock.Controller) (ArticleService, service.InteractiveService, repository.RankingRepository) {
				mockArticleService := svcmocks.NewMockArticleService(ctrl)
				mockInteractiveService := svcmocks.NewMockInteractiveService(ctrl)
				mockRankingRepository := repomocks.NewMockRankingRepository(ctrl)

				mockArticleService.EXPECT().ListPub(gomock.Any(), gomock.Any(), gomock.Any(), 0, 2).
					Return([]domain.Article{
						{Id: 1},
						{Id: 2},
					}, nil)
				mockArticleService.EXPECT().ListPub(gomock.Any(), gomock.Any(), gomock.Any(), 2, 2).
					Return([]domain.Article{
						{Id: 3},
						{Id: 4},
					}, nil)
				mockArticleService.EXPECT().ListPub(gomock.Any(), gomock.Any(), gomock.Any(), 4, 2).
					Return([]domain.Article{}, nil)

				mockInteractiveService.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2}).
					Return(map[int64]domain2.Interactive{
						1: {LikeCnt: 1},
						2: {LikeCnt: 2},
					}, nil)
				mockInteractiveService.EXPECT().GetByIds(gomock.Any(), "article", []int64{3, 4}).
					Return(map[int64]domain2.Interactive{
						3: {LikeCnt: 3},
						4: {LikeCnt: 4},
					}, nil)
				mockInteractiveService.EXPECT().GetByIds(gomock.Any(), "article", []int64{}).
					Return(map[int64]domain2.Interactive{}, nil)

				return mockArticleService, mockInteractiveService, mockRankingRepository
			},
			wantArts: []domain.Article{
				{Id: 4},
				{Id: 3},
				{Id: 2},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			artSvc, intraSvc, repo := tc.mock(controller)
			rankingSvc := &ArticleRankingService{repo: repo,
				artSvc:    artSvc,
				intraSvc:  intraSvc,
				before:    7 * 24 * time.Hour,
				n:         3,
				l:         InitialLogger(),
				batchSize: batchSize,
				bizStr:    "article",
				scoreFunc: func(likeCnt int64, utime time.Time) float64 {
					return float64(likeCnt)
				},
			}
			articles, err := rankingSvc.doRanking(context.Background())
			assert.Equal(t, tc.wantArts, articles)
			assert.Equal(t, tc.wantErr, err)
		})
	}

}

func InitialLogger() logger.LoggerV1 {
	cfg := zap.NewDevelopmentConfig()
	err := viper.UnmarshalKey("log", &cfg)
	if err != nil {
		panic(err)
	}
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}
