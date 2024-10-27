package code

import (
	"context"
	"errors"
	"testing"

	"github.com/misakimei123/redbook/internal/repository/cache/redismocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRedisCodeCache_Set(t *testing.T) {
	redisError := errors.New("redis fail")
	testCases := []struct {
		name      string
		ctx       context.Context
		biz       string
		phone     string
		code      string
		mock      func(controller *gomock.Controller) redis.Cmdable
		wantError error
	}{
		{
			name:  "set ok",
			ctx:   context.Background(),
			biz:   "",
			phone: "",
			code:  "",
			mock: func(controller *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(controller)
				result := redis.NewCmdResult(int64(0), nil)
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, gomock.Any(), gomock.Any()).Return(result)
				return cmd
			},
			wantError: nil,
		},
		{
			name:  "ErrCodeSendTooMany",
			ctx:   context.Background(),
			biz:   "",
			phone: "",
			code:  "",
			mock: func(controller *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(controller)
				result := redis.NewCmdResult(int64(-1), nil)
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, gomock.Any(), gomock.Any()).Return(result)
				return cmd
			},
			wantError: ErrCodeSendTooMany,
		},
		{
			name:  "RedisFail",
			ctx:   context.Background(),
			biz:   "",
			phone: "",
			code:  "",
			mock: func(controller *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(controller)
				result := redis.NewCmdResult(int64(-1), redisError)
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, gomock.Any(), gomock.Any()).Return(result)
				return cmd
			},
			wantError: redisError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			cmd := tc.mock(controller)
			cache := NewRedisCodeCache(cmd)
			err := cache.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantError, err)
		})
	}
}
