package integration

import (
	"context"
	"testing"
	"time"

	intrv1 "github.com/misakimei123/redbook/api/proto/gen/intr/v1"
	"github.com/misakimei123/redbook/interactive/grpc"
	"github.com/misakimei123/redbook/interactive/integration/startup"
	"github.com/misakimei123/redbook/interactive/repository/dao"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

func TestInteractiveService(t *testing.T) {
	suite.Run(t, &InteractiveTestSuite{})
}

type InteractiveTestSuite struct {
	suite.Suite
	db  *gorm.DB
	rdb redis.Cmdable
	svc *grpc.InteractiveServiceServer
}

func (s *InteractiveTestSuite) SetupSuite() {
	s.db = startup.InitDB()
	s.rdb = startup.InitRedis()
	s.svc = startup.InitInteractiveService()
}

func (s *InteractiveTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err := s.db.Exec("TRUNCATE TABLE `interactives`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `user_like_bizs`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `user_collect_bizs`").Error
	assert.NoError(s.T(), err)
	// 清空 Redis
	err = s.rdb.FlushDB(ctx).Err()
	assert.NoError(s.T(), err)
}

func (s *InteractiveTestSuite) TestIncrReadCnt() {
	test := s.T()
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		bizStr   string
		bizId    int64
		wantErr  error
		wantResp *intrv1.IncrReadCntResponse
	}{
		{
			name: "read cnt increase ok both in db and cache",
			before: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
				defer cancelFunc()
				now := time.Now().UnixMilli() - 1000
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         1,
					BizId:      1,
					BizStr:     "test",
					ReadCnt:    0,
					LikeCnt:    0,
					CollectCnt: 0,
					Utime:      0,
					Ctime:      now,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:1", "read_cnt", 0).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
				defer cancelFunc()
				var interactiveData dao.Interactive
				err := s.db.WithContext(ctx).Where("id=?", 1).First(&interactiveData).Error
				assert.NoError(t, err)
				assert.True(t, interactiveData.Utime > 0)
				interactiveData.Ctime = 0
				interactiveData.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:         1,
					BizId:      1,
					BizStr:     "test",
					ReadCnt:    1,
					LikeCnt:    0,
					CollectCnt: 0,
				}, interactiveData)
				cnt, err := s.rdb.HGet(ctx, "interactive:test:1", "read_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 1, cnt)
				err = s.rdb.Del(ctx, "interactive:test:1").Err()
				assert.NoError(t, err)
			},
			bizStr:   "test",
			bizId:    1,
			wantErr:  nil,
			wantResp: &intrv1.IncrReadCntResponse{},
		},
		{
			name: "read cnt increase only in db",
			before: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
				defer cancelFunc()
				now := time.Now().UnixMilli() - 1000
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         1,
					BizId:      1,
					BizStr:     "test",
					ReadCnt:    0,
					LikeCnt:    0,
					CollectCnt: 0,
					Utime:      0,
					Ctime:      now,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
				defer cancelFunc()
				var interactiveData dao.Interactive
				err := s.db.WithContext(ctx).Where("id=?", 1).First(&interactiveData).Error
				assert.NoError(t, err)
				assert.True(t, interactiveData.Utime > 0)
				interactiveData.Ctime = 0
				interactiveData.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:         1,
					BizId:      1,
					BizStr:     "test",
					ReadCnt:    1,
					LikeCnt:    0,
					CollectCnt: 0,
				}, interactiveData)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:1").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},
			bizStr:   "test",
			bizId:    1,
			wantErr:  nil,
			wantResp: &intrv1.IncrReadCntResponse{},
		},
		{
			name: "read cnt created in db",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
				defer cancelFunc()
				var interactiveData dao.Interactive
				err := s.db.WithContext(ctx).Where("id=?", 1).First(&interactiveData).Error
				assert.NoError(t, err)
				assert.True(t, interactiveData.Utime > 0)
				interactiveData.Ctime = 0
				interactiveData.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:         1,
					BizId:      1,
					BizStr:     "test",
					ReadCnt:    1,
					LikeCnt:    0,
					CollectCnt: 0,
				}, interactiveData)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:1").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)
			},
			bizStr:   "test",
			bizId:    1,
			wantErr:  nil,
			wantResp: &intrv1.IncrReadCntResponse{},
		},
	}
	for _, tc := range testCases {
		test.Run(tc.name, func(t *testing.T) {
			err := s.db.Exec("TRUNCATE TABLE `interactives`").Error
			assert.NoError(t, err)
			tc.before(t)
			resp, err := s.svc.IncrReadCnt(context.Background(), &intrv1.IncrReadCntRequest{
				BizStr: tc.bizStr,
				BizId:  tc.bizId,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
			tc.after(t)
		})
	}
}

func (s *InteractiveTestSuite) TestLike() {
	test := s.T()
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		like     bool
		bizStr   string
		bizId    int64
		uid      int64
		wantErr  error
		wantResp *intrv1.LikeResponse
	}{
		{
			name: "increase like in db and cache",
			before: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
				defer cancelFunc()
				now := time.Now().UnixMilli() - 1000
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         1,
					BizId:      1,
					BizStr:     "test",
					ReadCnt:    0,
					LikeCnt:    1,
					CollectCnt: 0,
					Utime:      0,
					Ctime:      now,
				}).Error
				s.db.WithContext(ctx).Create(dao.UserLikeBiz{
					Id:     1,
					Uid:    123,
					BizId:  1,
					BizStr: "test",
					Status: 1,
					Utime:  now,
					Ctime:  now,
				})
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:1", "like_cnt", 1).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
				defer cancelFunc()
				var uLike dao.UserLikeBiz
				err := s.db.WithContext(ctx).Where("uid=? and biz_id=? and biz_str=?", 123, 1, "test").First(&uLike).Error
				assert.NoError(t, err)
				assert.Equal(t, 1, uLike.Status)
				var interactiveData dao.Interactive
				err = s.db.WithContext(ctx).Where("id=?", 1).First(&interactiveData).Error
				assert.NoError(t, err)
				assert.True(t, interactiveData.Utime > 0)
				interactiveData.Ctime = 0
				interactiveData.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:         1,
					BizId:      1,
					BizStr:     "test",
					ReadCnt:    0,
					LikeCnt:    2,
					CollectCnt: 0,
				}, interactiveData)
				cnt, err := s.rdb.HGet(ctx, "interactive:test:1", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 2, cnt)
				err = s.rdb.Del(ctx, "interactive:test:1").Err()
				assert.NoError(t, err)
			},
			like:     true,
			bizStr:   "test",
			bizId:    1,
			uid:      123,
			wantErr:  nil,
			wantResp: &intrv1.LikeResponse{},
		},
		{
			name: "decrease like in db and cache",
			before: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
				defer cancelFunc()
				now := time.Now().UnixMilli() - 1000
				err := s.db.WithContext(ctx).Create(dao.Interactive{
					Id:         1,
					BizId:      1,
					BizStr:     "test",
					ReadCnt:    0,
					LikeCnt:    1,
					CollectCnt: 0,
					Utime:      0,
					Ctime:      now,
				}).Error
				s.db.WithContext(ctx).Create(dao.UserLikeBiz{
					Id:     1,
					Uid:    123,
					BizId:  1,
					BizStr: "test",
					Status: 1,
					Utime:  now,
					Ctime:  now,
				})
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:1", "like_cnt", 1).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
				defer cancelFunc()
				var uLike dao.UserLikeBiz
				err := s.db.WithContext(ctx).Where("uid=? and biz_id=? and biz_str=?", 123, 1, "test").First(&uLike).Error
				assert.NoError(t, err)
				assert.Equal(t, 0, uLike.Status)
				var interactiveData dao.Interactive
				err = s.db.WithContext(ctx).Where("id=?", 1).First(&interactiveData).Error
				assert.NoError(t, err)
				assert.True(t, interactiveData.Utime > 0)
				interactiveData.Ctime = 0
				interactiveData.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:         1,
					BizId:      1,
					BizStr:     "test",
					ReadCnt:    0,
					LikeCnt:    0,
					CollectCnt: 0,
				}, interactiveData)
				cnt, err := s.rdb.HGet(ctx, "interactive:test:1", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 0, cnt)
				err = s.rdb.Del(ctx, "interactive:test:1").Err()
				assert.NoError(t, err)
			},
			like:     false,
			bizStr:   "test",
			bizId:    1,
			uid:      123,
			wantErr:  nil,
			wantResp: &intrv1.LikeResponse{},
		},
		{
			name: "create like in db",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
				defer cancelFunc()
				var uLike dao.UserLikeBiz
				err := s.db.WithContext(ctx).
					Where("uid=? and biz_id=? and biz_str=?", 123, 1, "test").
					First(&uLike).Error
				assert.NoError(t, err)
				assert.Equal(t, 1, uLike.Status)
				var interactiveData dao.Interactive
				err = s.db.WithContext(ctx).Where("id=?", 1).First(&interactiveData).Error
				assert.NoError(t, err)
				assert.True(t, interactiveData.Utime > 0)
				interactiveData.Ctime = 0
				interactiveData.Utime = 0
				assert.Equal(t, dao.Interactive{
					Id:         1,
					BizId:      1,
					BizStr:     "test",
					ReadCnt:    0,
					LikeCnt:    1,
					CollectCnt: 0,
				}, interactiveData)
				result, err := s.rdb.Exists(ctx, "interactive:test:1").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), result)
				err = s.rdb.Del(ctx, "interactive:test:1").Err()
				assert.NoError(t, err)
			},
			like:     true,
			bizStr:   "test",
			bizId:    1,
			uid:      123,
			wantErr:  nil,
			wantResp: &intrv1.LikeResponse{},
		},
	}

	for _, tc := range testCases {
		test.Run(tc.name, func(t *testing.T) {
			err := s.db.Exec("TRUNCATE TABLE `interactives`").Error
			assert.NoError(s.T(), err)
			err = s.db.Exec("TRUNCATE TABLE `user_like_bizs`").Error
			assert.NoError(s.T(), err)
			tc.before(t)
			response, err := s.svc.Like(context.Background(), &intrv1.LikeRequest{
				BizStr: tc.bizStr,
				BizId:  tc.bizId,
				Uid:    tc.uid,
				Like:   tc.like,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, response)
			tc.after(t)
		})
	}
}

func (s *InteractiveTestSuite) TestCollect() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64
		cid   int64
		uid   int64

		wantErr  error
		wantResp *intrv1.CollectResponse
	}{
		{
			name: "collect ok, not in db and buffer",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
				defer cancelFunc()
				var intr dao.Interactive
				err := s.db.Where("biz_str = ? and biz_id = ?", "test", 1).First(&intr).Error
				assert.NoError(t, err)
				assert.True(t, intr.Ctime > 0)
				intr.Ctime = 0
				assert.True(t, intr.Utime > 0)
				intr.Utime = 0
				assert.True(t, intr.Id > 0)
				intr.Id = 0
				assert.Equal(t, dao.Interactive{
					BizStr:     "test",
					BizId:      1,
					CollectCnt: 1,
				}, intr)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:1").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)

				var cbiz dao.UserCollectBiz
				err = s.db.WithContext(ctx).
					Where("uid = ? and biz_str = ? and biz_id = ?", 1, "test", 1).First(&cbiz).Error
				assert.NoError(t, err)
				assert.True(t, cbiz.Ctime > 0)
				cbiz.Ctime = 0
				assert.True(t, cbiz.Utime > 0)
				cbiz.Utime = 0
				assert.True(t, cbiz.Id > 0)
				cbiz.Id = 0
				assert.Equal(t, dao.UserCollectBiz{
					BizStr: "test",
					BizId:  1,
					Cid:    1,
					Uid:    1,
				}, cbiz)
			},
			biz:      "test",
			bizId:    1,
			cid:      1,
			uid:      1,
			wantErr:  nil,
			wantResp: &intrv1.CollectResponse{},
		},
		{
			name: "collect ok, in db not in rdb",
			before: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
				defer cancelFunc()
				err := s.db.WithContext(ctx).Create(&dao.Interactive{
					BizId:      2,
					BizStr:     "test",
					CollectCnt: 10,
					Ctime:      123,
					Utime:      234,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
				defer cancelFunc()
				var intr dao.Interactive
				err := s.db.WithContext(ctx).Where("biz_str = ? and biz_id = ?", "test", 2).First(&intr).Error
				assert.NoError(t, err)
				assert.True(t, intr.Ctime > 0)
				intr.Ctime = 0
				assert.True(t, intr.Utime > 0)
				intr.Utime = 0
				assert.True(t, intr.Id > 0)
				intr.Id = 0
				assert.Equal(t, dao.Interactive{
					BizStr:     "test",
					BizId:      2,
					CollectCnt: 11,
				}, intr)
				cnt, err := s.rdb.Exists(ctx, "interactive:test:2").Result()
				assert.NoError(t, err)
				assert.Equal(t, int64(0), cnt)

				var cbiz dao.UserCollectBiz
				err = s.db.WithContext(ctx).
					Where("uid = ? and biz_str = ? and biz_id = ?", 1, "test", 2).First(&cbiz).Error
				assert.NoError(t, err)
				assert.True(t, cbiz.Ctime > 0)
				cbiz.Ctime = 0
				assert.True(t, cbiz.Utime > 0)
				cbiz.Utime = 0
				assert.True(t, cbiz.Id > 0)
				cbiz.Id = 0
				assert.Equal(t, dao.UserCollectBiz{
					BizStr: "test",
					BizId:  2,
					Cid:    1,
					Uid:    1,
				}, cbiz)
			},
			biz:      "test",
			bizId:    2,
			cid:      1,
			uid:      1,
			wantResp: &intrv1.CollectResponse{},
		},
		{
			name: "collect success, in db and cache",
			before: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
				defer cancelFunc()
				err := s.db.WithContext(ctx).Create(&dao.Interactive{
					BizStr:     "test",
					BizId:      3,
					CollectCnt: 10,
					Ctime:      123,
					Utime:      234,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:3", "collect_cnt", 10).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
				defer cancelFunc()
				var intr dao.Interactive
				err := s.db.WithContext(ctx).Where("biz_str = ? and biz_id = ?", "test", 3).First(&intr).Error
				assert.NoError(t, err)
				assert.True(t, intr.Ctime > 0)
				intr.Ctime = 0
				assert.True(t, intr.Utime > 0)
				intr.Utime = 0
				assert.True(t, intr.Id > 0)
				intr.Id = 0
				assert.Equal(t, dao.Interactive{
					BizStr:     "test",
					BizId:      3,
					CollectCnt: 11,
				}, intr)
				cnt, err := s.rdb.HGet(ctx, "interactive:test:3", "collect_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 11, cnt)

				var cbiz dao.UserCollectBiz
				err = s.db.WithContext(ctx).Where("uid = ? and biz_str = ? and biz_id = ?", 1, "test", 3).
					First(&cbiz).Error
				assert.NoError(t, err)
				assert.True(t, cbiz.Ctime > 0)
				cbiz.Ctime = 0
				assert.True(t, cbiz.Utime > 0)
				cbiz.Utime = 0
				assert.True(t, cbiz.Id > 0)
				cbiz.Id = 0
				assert.Equal(t, dao.UserCollectBiz{
					BizStr: "test",
					BizId:  3,
					Cid:    1,
					Uid:    1,
				}, cbiz)
			},
			bizId:    3,
			biz:      "test",
			cid:      1,
			uid:      1,
			wantResp: &intrv1.CollectResponse{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := s.db.Exec("TRUNCATE TABLE `interactives`").Error
			assert.NoError(s.T(), err)
			err = s.db.Exec("TRUNCATE TABLE `user_collect_bizs`").Error
			assert.NoError(s.T(), err)
			tc.before(t)
			resp, err := s.svc.Collect(context.Background(), &intrv1.CollectRequest{
				BizStr: tc.biz,
				BizId:  tc.bizId,
				Uid:    tc.uid,
				Cid:    tc.cid,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
			tc.after(t)
		})

	}
}

func (s *InteractiveTestSuite) TestGet() {
	t := s.T()
	testCases := []struct {
		name    string
		before  func(t *testing.T)
		bizId   int64
		biz     string
		uid     int64
		wantErr error
		wantRes *intrv1.GetResponse
	}{
		{
			name:  "get all, no buffer",
			bizId: 12,
			biz:   "test",
			uid:   123,
			before: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*3)
				defer cancelFunc()
				err := s.db.WithContext(ctx).Create(&dao.Interactive{
					BizId:      12,
					BizStr:     "test",
					ReadCnt:    500,
					LikeCnt:    200,
					CollectCnt: 300,
					Utime:      200,
					Ctime:      100,
				}).Error
				assert.NoError(t, err)
			},
			wantRes: &intrv1.GetResponse{Interactive: &intrv1.Interactive{
				BizStr:     "test",
				BizId:      12,
				ReadCnt:    500,
				LikeCnt:    200,
				CollectCnt: 300,
			}},
		},
		{
			name:  "get all, buffered and liked collected by user",
			biz:   "test",
			bizId: 3,
			uid:   123,
			before: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*3)
				defer cancelFunc()
				err := s.db.WithContext(ctx).Create(&dao.UserCollectBiz{
					Uid:    123,
					BizId:  3,
					BizStr: "test",
					Cid:    1,
					Utime:  200,
					Ctime:  100,
				}).Error
				assert.NoError(t, err)
				err = s.db.WithContext(ctx).Create(&dao.UserLikeBiz{
					Uid:    123,
					BizId:  3,
					BizStr: "test",
					Status: 1,
					Utime:  200,
					Ctime:  100,
				}).Error
				assert.NoError(t, err)
				err = s.rdb.HSet(ctx, "interactive:test:3", "read_cnt", 1, "collect_cnt", 1).Err()
				assert.NoError(t, err)
			},
			wantRes: &intrv1.GetResponse{Interactive: &intrv1.Interactive{
				BizStr:     "test",
				BizId:      3,
				ReadCnt:    1,
				CollectCnt: 1,
				Liked:      true,
				Collected:  true,
			}},
			//wantRes: domain.Interactive{
			//	BizId:      3,
			//	Biz:        "test",
			//	ReadCnt:    1,
			//	CollectCnt: 1,
			//	Liked:      true,
			//	Collected:  true,
			//},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			res, err := s.svc.Get(context.Background(), &intrv1.GetRequest{
				BizStr: tc.biz,
				BizId:  tc.bizId,
				Uid:    tc.uid,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantRes, res)
			err = s.db.Exec("TRUNCATE TABLE `interactives`").Error
			assert.NoError(s.T(), err)
			err = s.db.Exec("TRUNCATE TABLE `user_like_bizs`").Error
			assert.NoError(s.T(), err)
			err = s.db.Exec("TRUNCATE TABLE `user_collect_bizs`").Error
			assert.NoError(s.T(), err)
		})
	}
}

func (s *InteractiveTestSuite) TestGetByIds() {
	t := s.T()
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*3)
	defer cancelFunc()

	for i := 1; i < 5; i++ {
		i := int64(i)
		err := s.db.WithContext(ctx).Create(&dao.Interactive{
			Id:         i,
			BizId:      i,
			BizStr:     "test",
			ReadCnt:    i + 1,
			LikeCnt:    i,
			CollectCnt: i,
		}).Error
		assert.NoError(t, err)
	}

	testCases := []struct {
		name    string
		biz     string
		ids     []int64
		wantErr error
		wantRes *intrv1.GetByIdsResponse
	}{
		{
			name: "query ok",
			biz:  "test",
			ids:  []int64{1, 2},
			wantRes: &intrv1.GetByIdsResponse{Interacs: map[int64]*intrv1.Interactive{
				1: {
					BizId:      1,
					ReadCnt:    2,
					LikeCnt:    1,
					CollectCnt: 1,
				},
				2: {
					BizId:      2,
					ReadCnt:    3,
					LikeCnt:    2,
					CollectCnt: 2,
				},
			}},
			//wantRes: map[int64]domain.Interactive{
			//	1: {
			//		BizId:      1,
			//		ReadCnt:    2,
			//		LikeCnt:    1,
			//		CollectCnt: 1,
			//	},
			//	2: {
			//		BizId:      2,
			//		ReadCnt:    3,
			//		LikeCnt:    2,
			//		CollectCnt: 2,
			//	},
			//},
		},
		{
			name:    "query fail",
			biz:     "test",
			ids:     []int64{10, 20},
			wantRes: &intrv1.GetByIdsResponse{Interacs: map[int64]*intrv1.Interactive{}},
		},
	}
	svc := startup.InitInteractiveService()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := svc.GetByIds(context.Background(), &intrv1.GetByIdsRequest{
				BizStr: tc.biz,
				Ids:    tc.ids,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
