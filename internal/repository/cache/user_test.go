package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/repository/cache/redismocks"
	"go.uber.org/mock/gomock"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisUserCache_Set(t *testing.T) {
	user := domain.User{
		Id:       0,
		Email:    "123@123.net",
		Password: "123",
		Nick:     "123",
		AboutMe:  "123",
		Birthday: time.Now(),
		Phone:    "123",
	}

	key := fmt.Sprintf("user:info:%d", user.Id)
	userJson, _ := json.Marshal(user)

	testCases := []struct {
		name      string
		ctx       context.Context
		du        domain.User
		mock      func(controller *gomock.Controller) redis.Cmdable
		wantError error
	}{
		{
			name: "set ok",
			ctx:  context.Background(),
			du:   user,
			mock: func(controller *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(controller)
				cmdResult := redis.NewStatusCmd(context.Background(), nil)
				cmd.EXPECT().Set(gomock.Any(), key, userJson, gomock.Any()).Return(cmdResult)
				return cmd
			},
			wantError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			redisClient := tc.mock(controller)
			cache := NewRedisUserCache(redisClient)
			err := cache.Set(tc.ctx, tc.du)
			assert.Equal(t, tc.wantError, err)
		})
	}

}
