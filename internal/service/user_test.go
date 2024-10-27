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
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_Login(t *testing.T) {
	const (
		password = "hello#world123"
		email    = "123@qq.com"
	)
	dbError := errors.New("db fail")
	encryptPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := domain.User{
		Id:       0,
		Email:    email,
		Password: string(encryptPassword),
		Phone:    "12123456789",
	}

	testCases := []struct {
		name      string
		ctx       context.Context
		email     string
		password  string
		mock      func(ctrl *gomock.Controller) repository.UserRepository
		wantUser  domain.User
		wantError error
	}{
		{
			name:     "登录成功",
			ctx:      context.Background(),
			email:    email,
			password: password,
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				mockUserRepository := repomocks.NewMockUserRepository(ctrl)
				mockUserRepository.EXPECT().FindByEmail(gomock.Any(), email).Return(user, nil)
				return mockUserRepository
			},
			wantUser:  user,
			wantError: nil,
		},
		{
			name:     "user not found",
			ctx:      context.Background(),
			email:    email,
			password: password,
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				mockUserRepository := repomocks.NewMockUserRepository(ctrl)
				mockUserRepository.EXPECT().FindByEmail(gomock.Any(), email).Return(user, repository.ErrUserNotFound)
				return mockUserRepository
			},
			//wantUser:  domain.User{},
			wantError: ErrInvalidUserOrPassword,
		},
		{
			name:     "db error",
			ctx:      context.Background(),
			email:    email,
			password: password,
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				mockUserRepository := repomocks.NewMockUserRepository(ctrl)
				mockUserRepository.EXPECT().FindByEmail(gomock.Any(), email).Return(user, dbError)
				return mockUserRepository
			},
			//wantUser:  domain.User{},
			wantError: dbError,
		},
		{
			name:     "password not match",
			ctx:      context.Background(),
			email:    email,
			password: password + "1",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				mockUserRepository := repomocks.NewMockUserRepository(ctrl)
				mockUserRepository.EXPECT().FindByEmail(gomock.Any(), email).Return(user, nil)
				return mockUserRepository
			},
			//wantUser:  domain.User{},
			wantError: ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()
			userRepo := tc.mock(controller)
			userService := NewUserService(userRepo)
			user, err := userService.Login(tc.ctx, tc.email, tc.password)
			assert.Equal(t, tc.wantError, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
