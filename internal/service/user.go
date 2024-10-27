package service

import (
	"context"
	"errors"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/repository"
	"github.com/misakimei123/redbook/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("invalid user or password")
)

type UserService interface {
	SignUp(ctx context.Context, user *domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	Edit(ctx context.Context, user domain.User) error
	Profile(ctx context.Context, userId int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
	l    logger.LoggerV1
}

func (s *userService) FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error) {
	user, err := s.repo.FindByWechat(ctx, wechatInfo.OpenId)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return user, err
	}
	user = domain.User{
		WechatInfo: wechatInfo,
	}
	err = s.repo.Create(ctx, &user)
	if err != nil && !errors.Is(err, repository.ErrDuplicateUser) {
		return user, err
	}
	return s.repo.FindByWechat(ctx, wechatInfo.OpenId)
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) SignUp(ctx context.Context, user *domain.User) error {
	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(password)
	err = s.repo.Create(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func (s *userService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return user, nil
}

func (s *userService) Edit(ctx context.Context, user domain.User) error {
	err := s.repo.UpdateProfile(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func (s *userService) Profile(ctx context.Context, userId int64) (domain.User, error) {
	user, err := s.repo.FindByID(ctx, userId)
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	user, err := s.repo.FindByPhone(ctx, phone)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return user, err
	}
	user = domain.User{
		Phone: phone,
	}
	err = s.repo.Create(ctx, &user)
	if err != nil && !errors.Is(err, repository.ErrDuplicateUser) {
		return user, err
	}
	// 这里看看create返回后有没有回填user
	return s.repo.FindByPhone(ctx, phone)
}
