package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/integration/startup"
	"github.com/misakimei123/redbook/internal/repository/dao"
	"github.com/misakimei123/redbook/internal/web/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

func TestArticleGormHandler(t *testing.T) {
	suite.Run(t, &ArticleHandlerSuite{})
}

type ArticleHandlerSuite struct {
	suite.Suite
	db     *gorm.DB
	server *gin.Engine
}

func (s *ArticleHandlerSuite) SetupSuite() {
	s.db = startup.InitDB()
	articleHandler := startup.InitArticleHandler(dao.NewArticleGormDao(s.db))
	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", jwt.UserClaims{Uid: 123})
	})
	articleHandler.RegisterRoutes(server)
	s.server = server
}

func (s *ArticleHandlerSuite) TearDownTest() {
	s.db.Exec("truncate table `articles`")
	s.db.Exec("truncate table `published_articles`")
}

func (s *ArticleHandlerSuite) TestEdit() {
	t := s.T()
	testCases := []testCase{
		{
			name: "create new article",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				// check db
				var art dao.Article
				err := s.db.Where("author_id=?", 123).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Ctime > 0)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: int64(123),
					Status:   domain.ArticleStatusUnpublished,
				}, art)
			},
			art: Article{
				Title:   "my article",
				Content: "my article content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 0,
				Msg:  "",
				Data: 1,
			},
		},
		{
			name: "update article",
			before: func(t *testing.T) {
				s.db.Create(dao.Article{
					Id:       2,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: 123,
					Ctime:    123,
					Utime:    123,
				})
			},
			after: func(t *testing.T) {
				// check db
				var art dao.Article
				err := s.db.Where("id=?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 123)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "new article",
					Content:  "new article content",
					AuthorId: int64(123),
					Status:   domain.ArticleStatusUnpublished,
				}, art)
			},
			art: Article{
				Id:      2,
				Title:   "new article",
				Content: "new article content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 0,
				Msg:  "",
				Data: 2,
			},
		},
		{
			name: "update other's article",
			before: func(t *testing.T) {
				s.db.Create(dao.Article{
					Id:       3,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: 456,
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    123,
					Utime:    456,
				})
			},
			after: func(t *testing.T) {
				// check db
				var art dao.Article
				err := s.db.Where("id=?", 3).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: int64(456),
					Ctime:    int64(123),
					Utime:    int64(456),
					Status:   domain.ArticleStatusUnpublished,
				}, art)
			},
			art: Article{
				Id:      3,
				Title:   "new article",
				Content: "new article content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "system error",
				Data: 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			body, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			request, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, request)
			assert.Equal(t, tc.wantCode, recorder.Code)
			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func (s *ArticleHandlerSuite) TestPublish() {
	t := s.T()
	testCases := []testCase{
		{
			name: "create and publish success",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				var art dao.Article
				err := s.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				now := time.Now().UnixMilli() - 3600*1000
				assert.True(t, art.Utime > now)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "my title",
					Content:  "my content",
					AuthorId: int64(123),
					Status:   domain.ArticleStatusPublished,
				}, art)
				var pubArt dao.PublishedArticle
				err = s.db.Where("id=?", 1).First(&pubArt).Error
				assert.NoError(t, err)
				now = time.Now().UnixMilli() - 3600*1000
				assert.True(t, pubArt.Utime > now)
				pubArt.Ctime = 0
				pubArt.Utime = 0
				assert.Equal(t, dao.PublishedArticle{
					Id:       1,
					Title:    "my title",
					Content:  "my content",
					AuthorId: int64(123),
					Status:   domain.ArticleStatusPublished,
				}, pubArt)
			},
			art: Article{
				Title:   "my title",
				Content: "my content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 1,
			},
		},
		{
			name: "modify and publish success",
			before: func(t *testing.T) {
				now := time.Now().UnixMilli()
				s.db.Create(dao.Article{
					Id:       2,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: 123,
					Ctime:    now,
					Utime:    now,
				})
			},
			after: func(t *testing.T) {
				var art dao.Article
				err := s.db.Where("id=?", 2).First(&art).Error
				assert.NoError(t, err)
				now := time.Now().UnixMilli() - 10*1000
				assert.True(t, art.Utime > now)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "my title",
					Content:  "my content",
					AuthorId: int64(123),
					Status:   domain.ArticleStatusPublished,
				}, art)
				var pubArt dao.PublishedArticle
				err = s.db.Where("id=?", 2).First(&pubArt).Error
				assert.NoError(t, err)
				now = time.Now().UnixMilli() - 3*1000
				assert.True(t, pubArt.Utime > now)
				pubArt.Ctime = 0
				pubArt.Utime = 0
				assert.Equal(t, dao.PublishedArticle{
					Id:       2,
					Title:    "my title",
					Content:  "my content",
					AuthorId: int64(123),
					Status:   domain.ArticleStatusPublished,
				}, pubArt)
			},
			art: Article{
				Id:      2,
				Title:   "my title",
				Content: "my content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 2,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			body, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			request, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, request)
			assert.Equal(t, tc.wantCode, recorder.Code)
			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type testCase struct {
	name     string
	before   func(t *testing.T)
	after    func(t *testing.T)
	art      Article
	wantCode int
	wantRes  Result[int64]
}
