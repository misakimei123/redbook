package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/service"
	svcmocks "github.com/misakimei123/redbook/internal/service/mocks"
	"github.com/misakimei123/redbook/internal/web/jwt"
	"github.com/misakimei123/redbook/internal/web/result"
	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestNewArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(controller *gomock.Controller) service.ArticleService
		reqBody  string
		wantCode int
		wantRes  result.Result
	}{
		{
			name: "publish success",
			mock: func(controller *gomock.Controller) service.ArticleService {
				articleService := svcmocks.NewMockArticleService(controller)
				articleService.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author:  domain.Author{Id: 123},
				}).Return(int64(1), nil)
				return articleService
			},
			reqBody: `
{
    "title": "my title",
    "content": "my content"
}
`,
			wantCode: 200,
			wantRes:  result.Result{Data: float64(1)},
		},
		{
			name: "publish success for published article",
			mock: func(controller *gomock.Controller) service.ArticleService {
				articleService := svcmocks.NewMockArticleService(controller)
				articleService.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      123,
					Title:   "my title",
					Content: "my content",
					Author:  domain.Author{Id: 123},
				}).Return(int64(123), nil)
				return articleService
			},
			reqBody: `
{
	"id": 123,
    "title": "my title",
    "content": "my content"
}
`,
			wantCode: 200,
			wantRes:  result.Result{Data: float64(123)},
		},
		{
			name: "publish fail for published article",
			mock: func(controller *gomock.Controller) service.ArticleService {
				articleService := svcmocks.NewMockArticleService(controller)
				articleService.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      123,
					Title:   "my title",
					Content: "my content",
					Author:  domain.Author{Id: 123},
				}).Return(int64(0), errors.New("mock error"))
				return articleService
			},
			reqBody: `
{
	"id": 123,
    "title": "my title",
    "content": "my content"
}
`,
			wantCode: 200,
			wantRes: result.Result{
				Code: 501001,
				Msg:  "system error"},
		},
		{
			name: "bind error",
			mock: func(controller *gomock.Controller) service.ArticleService {
				articleService := svcmocks.NewMockArticleService(controller)
				return articleService
			},
			reqBody: `
{
	"id": 123,
    "title": "my title"111,
    "content": "my content"
}
`,
			wantCode: 400,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			articleService := tc.mock(ctrl)

			l, err := zap.NewDevelopment()
			if err != nil {
				t.Fatal(err)
			}

			handler := NewArticleHandler(articleService, logger.NewZapLogger(l))
			server := gin.Default()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user", jwt.UserClaims{Uid: 123})
			})
			handler.RegisterRoutes(server)
			request, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBufferString(tc.reqBody))
			request.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, request)
			assert.Equal(t, tc.wantCode, recorder.Code)
			if recorder.Code != http.StatusOK {
				return
			}
			var res result.Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
