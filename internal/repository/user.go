package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/repository/cache"
	"github.com/misakimei123/redbook/internal/repository/dao"
	"github.com/redis/go-redis/v9"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateUser
	ErrUserNotFound  = dao.ErrRecordNotFound
	ErrEmptyKey      = redis.Nil
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateProfile(ctx context.Context, user domain.User) error
	FindByID(ctx context.Context, userId int64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByWechat(ctx context.Context, openId string) (domain.User, error)
}

type CacheUserRepository struct {
	userDao    dao.UserDao
	cache      cache.UserCache
	profileDao dao.ProfileDao
}

func (r *CacheUserRepository) Create(ctx context.Context, user *domain.User) error {
	return r.userDao.Insert(ctx, r.toEntity(user))
}

func (r *CacheUserRepository) FindByID(ctx context.Context, userId int64) (domain.User, error) {
	du, err := r.cache.Get(ctx, userId)
	switch err {
	case nil:
		return du, nil
	case ErrEmptyKey:
		profile, err := r.profileDao.FindByUserId(ctx, userId)
		if err != nil {
			return domain.User{}, err
		}
		du = domain.User{
			Id:       userId,
			Nick:     profile.Nick.String,
			AboutMe:  profile.AboutMe.String,
			Birthday: time.Unix(int64(profile.Birthday.Int64/1000), 0),
		}
		err = r.cache.Set(ctx, du)
		if err != nil {
			return domain.User{}, err
		}
		return du, err
	default:
		return domain.User{}, err
	}
}

func (r *CacheUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	user, err := r.userDao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return toDomainUser(user), nil
}

func (r *CacheUserRepository) FindByWechat(ctx context.Context, openId string) (domain.User, error) {
	user, err := r.userDao.FindByWechat(ctx, openId)
	if err != nil {
		return domain.User{}, err
	}
	return toDomainUser(user), nil
}

func NewCacheUserRepository(userDao dao.UserDao, profileDao dao.ProfileDao, c cache.UserCache) UserRepository {
	return &CacheUserRepository{
		userDao:    userDao,
		profileDao: profileDao,
		cache:      c,
	}
}

func (r *CacheUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := r.userDao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return toDomainUser(user), nil
}

func (r *CacheUserRepository) UpdateProfile(ctx context.Context, user domain.User) error {
	return r.profileDao.Update(ctx, &dao.Profile{
		UserId: user.Id,
		User:   toDaoUser(user),
		Birthday: sql.NullInt64{
			Int64: user.Birthday.UnixMilli(),
			Valid: !user.Birthday.IsZero(),
		},
		Nick:    String2SqlNullString(user.Nick),
		AboutMe: String2SqlNullString(user.AboutMe),
	})
}

func (r CacheUserRepository) toEntity(u *domain.User) *dao.User {
	now := time.Now().UnixMilli()
	return &dao.User{
		Id:            u.Id,
		Email:         String2SqlNullString(u.Email),
		Password:      String2SqlNullString(u.Password),
		Phone:         String2SqlNullString(u.Phone),
		Ctime:         now,
		Utime:         now,
		WechatOpenId:  String2SqlNullString(u.WechatInfo.OpenId),
		WechatUnionId: String2SqlNullString(u.WechatInfo.UnionId),
	}
}

func toDaoUser(user domain.User) dao.User {
	return dao.User{
		Id: user.Id,
	}
}

func toDomainUser(user dao.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email.String,
		Password: user.Password.String,
		Phone:    user.Phone.String,
		WechatInfo: domain.WechatInfo{
			UnionId: user.WechatUnionId.String,
			OpenId:  user.WechatOpenId.String,
		},
	}
}

func String2SqlNullString(field string) sql.NullString {
	return sql.NullString{
		String: field,
		Valid:  len(field) > 0,
	}
}
