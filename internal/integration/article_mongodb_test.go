package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/integration/startup"
	"github.com/misakimei123/redbook/internal/repository/dao"
	"github.com/misakimei123/redbook/internal/web/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestArticleMongoDBHandler(t *testing.T) {
	suite.Run(t, &ArticleMongoDBHandlerSuite{})
}

type ArticleMongoDBHandlerSuite struct {
	suite.Suite
	server  *gin.Engine
	db      *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (s *ArticleMongoDBHandlerSuite) SetupSuite() {
	s.db = startup.InitMongoDB()
	s.col = s.db.Collection("articles")
	s.liveCol = s.db.Collection("published_articles")
	node, err := snowflake.NewNode(1)
	assert.NoError(s.T(), err)
	articleHandler := startup.InitArticleHandler(dao.NewMongoDBArticleDAO(s.db, node))
	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", jwt.UserClaims{Uid: 123})
	})
	articleHandler.RegisterRoutes(server)
	s.server = server
}

func (s *ArticleMongoDBHandlerSuite) TearDownTest() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	_, err := s.col.DeleteMany(ctx, bson.D{})
	t := s.T()
	assert.NoError(t, err)
	_, err = s.liveCol.DeleteMany(ctx, bson.D{})
	assert.NoError(t, err)
}

func (s *ArticleMongoDBHandlerSuite) TestEdit() {
	t := s.T()
	testCases := []testCase{
		{
			name: "create new article",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				// check db
				ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancelFunc()
				filter := bson.D{bson.E{"author_id", 123}}
				result := s.col.FindOne(ctx, filter)
				assert.NoError(t, result.Err())
				var art dao.Article
				err := result.Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Ctime > 0)
				art.Ctime = 0
				art.Utime = 0
				art.Id = 1
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
				ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancelFunc()
				_, err := s.col.InsertOne(ctx, dao.Article{
					Id:       2,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: 123,
					Ctime:    123,
					Utime:    123,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// check db
				ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancelFunc()
				var art dao.Article
				filter := bson.D{bson.E{"id", 2}}
				err := s.col.FindOne(ctx, filter).Decode(&art)
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
				ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancelFunc()
				_, err := s.col.InsertOne(ctx, dao.Article{
					Id:       3,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: 456,
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    123,
					Utime:    456,
				})
				if err != nil {
					return
				}
			},
			after: func(t *testing.T) {
				// check db
				ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancelFunc()
				var art dao.Article
				filter := bson.D{bson.E{"id", 3}}
				err := s.col.FindOne(ctx, filter).Decode(&art)
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
			if tc.wantRes.Data > 0 {
				assert.True(t, res.Data > 0)
			}
		})
	}
}

func (s *ArticleMongoDBHandlerSuite) TestPublish() {
	t := s.T()
	testCases := []testCase{
		{
			name: "create and publish success",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancelFunc()
				var art dao.Article
				filter := bson.D{bson.E{"author_id", 123}}
				err := s.col.FindOne(ctx, filter).Decode(&art)
				assert.NoError(t, err)
				now := time.Now().UnixMilli() - 3600*1000
				assert.True(t, art.Utime > now)
				art.Ctime = 0
				art.Utime = 0
				art.Id = 1
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "my title",
					Content:  "my content",
					AuthorId: int64(123),
					Status:   domain.ArticleStatusPublished,
				}, art)
				var pubArt dao.PublishedArticle
				err = s.liveCol.FindOne(ctx, filter).Decode(&pubArt)
				assert.NoError(t, err)
				now = time.Now().UnixMilli() - 3600*1000
				assert.True(t, pubArt.Utime > now)
				pubArt.Ctime = 0
				pubArt.Utime = 0
				pubArt.Id = 1
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
				ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancelFunc()
				now := time.Now().UnixMilli()
				_, err := s.col.InsertOne(ctx, dao.Article{
					Id:       2,
					Title:    "my article",
					Content:  "my article content",
					AuthorId: 123,
					Ctime:    now,
					Utime:    now,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancelFunc()
				var art dao.Article
				filter := bson.D{bson.E{"id", 2}}
				err := s.col.FindOne(ctx, filter).Decode(&art)
				assert.NoError(t, err)
				now := time.Now().UnixMilli() - 1*1000
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
				filter = bson.D{bson.E{"id", 2}}
				err = s.liveCol.FindOne(ctx, filter).Decode(&pubArt)
				assert.NoError(t, err)
				now = time.Now().UnixMilli() - 1*1000
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
			//assert.Equal(t, tc.wantRes, res)
			if tc.wantRes.Data > 0 {
				assert.True(t, res.Data > 0)
			}
		})
	}
}
