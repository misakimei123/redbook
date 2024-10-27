package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/repository/cache"
	cachemocks "github.com/misakimei123/redbook/internal/repository/cache/mocks"
	"github.com/misakimei123/redbook/internal/repository/dao"
	daomocks "github.com/misakimei123/redbook/internal/repository/dao/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCacheUserRepository_FindByID(t *testing.T) {
	const (
		userId = int64(0)
	)
	emptyDu := domain.User{}
	now := time.Now()
	profile := dao.Profile{
		UserId: 0,
		Birthday: sql.NullInt64{
			Int64: now.UnixMilli(),
			Valid: true,
		},
		Nick: sql.NullString{
			String: "",
			Valid:  true,
		},
		AboutMe: sql.NullString{
			String: "",
			Valid:  true,
		},
	}
	du := domain.User{
		Id:       0,
		Nick:     "",
		AboutMe:  "",
		Birthday: time.Unix(int64(profile.Birthday.Int64/1000), 0),
	}
	redisError := errors.New("redis error")

	testCases := []struct {
		name      string
		ctx       context.Context
		userId    int64
		mock      func(controller *gomock.Controller) (dao.UserDao, dao.ProfileDao, cache.UserCache)
		wantUser  domain.User
		wantError error
	}{
		{
			name:   "not in cache",
			ctx:    context.Background(),
			userId: 0,
			mock: func(controller *gomock.Controller) (dao.UserDao, dao.ProfileDao, cache.UserCache) {
				userDao := daomocks.NewMockUserDao(controller)
				profileDao := daomocks.NewMockProfileDao(controller)
				userCache := cachemocks.NewMockUserCache(controller)
				userCache.EXPECT().Get(gomock.Any(), userId).Return(emptyDu, ErrEmptyKey)
				profileDao.EXPECT().FindByUserId(gomock.Any(), userId).Return(profile, nil)
				userCache.EXPECT().Set(gomock.Any(), du).Return(nil)
				return userDao, profileDao, userCache
			},
			wantUser:  du,
			wantError: nil,
		},
		{
			name:   "in cache",
			ctx:    context.Background(),
			userId: 0,
			mock: func(controller *gomock.Controller) (dao.UserDao, dao.ProfileDao, cache.UserCache) {
				userDao := daomocks.NewMockUserDao(controller)
				profileDao := daomocks.NewMockProfileDao(controller)
				userCache := cachemocks.NewMockUserCache(controller)
				userCache.EXPECT().Get(gomock.Any(), userId).Return(du, nil)
				return userDao, profileDao, userCache
			},
			wantUser:  du,
			wantError: nil,
		},
		{
			name:   "cache error",
			ctx:    context.Background(),
			userId: 0,
			mock: func(controller *gomock.Controller) (dao.UserDao, dao.ProfileDao, cache.UserCache) {
				userDao := daomocks.NewMockUserDao(controller)
				profileDao := daomocks.NewMockProfileDao(controller)
				userCache := cachemocks.NewMockUserCache(controller)
				userCache.EXPECT().Get(gomock.Any(), userId).Return(emptyDu, redisError)
				return userDao, profileDao, userCache
			},
			wantUser:  emptyDu,
			wantError: redisError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			userDao, profileDao, userCache := tc.mock(controller)
			userRepository := NewCacheUserRepository(userDao, profileDao, userCache)
			user, err := userRepository.FindByID(tc.ctx, tc.userId)
			assert.Equal(t, tc.wantUser, user)
			assert.Equal(t, tc.wantError, err)
		})
	}
}
