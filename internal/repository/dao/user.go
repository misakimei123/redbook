package dao

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrDuplicateUser  = errors.New("email duplicated")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDao interface {
	Insert(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByWechat(ctx context.Context, openId string) (User, error)
}

type GormUserDao struct {
	db *gorm.DB
}

func NewGormUserDao(db *gorm.DB) UserDao {
	return &GormUserDao{db: db}
}

type User struct {
	Id            int64          `gorm:"primaryKey, autoIncrement"`
	Email         sql.NullString `gorm:"unique"`
	Password      sql.NullString
	Ctime         int64
	Utime         int64
	Phone         sql.NullString `gorm:"unique"`
	WechatOpenId  sql.NullString `gorm:"unique"`
	WechatUnionId sql.NullString
}

func (g *GormUserDao) Insert(ctx context.Context, user *User) error {
	err := g.db.WithContext(ctx).Create(user).Error
	if sqlError, ok := err.(*mysql.MySQLError); ok {
		const duplicateErr uint16 = 1062
		if sqlError.Number == duplicateErr {
			return ErrDuplicateUser
		}
	}

	return err
}

func (g *GormUserDao) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := g.db.WithContext(ctx).Where("email=?", email).First(&user).Error
	return user, err
}

func (g *GormUserDao) FindByPhone(ctx context.Context, phone string) (User, error) {
	var user User
	err := g.db.WithContext(ctx).Where("phone=?", phone).First(&user).Error
	return user, err
}

func (d *GormUserDao) FindByWechat(ctx context.Context, openId string) (User, error) {
	var user User
	err := d.db.WithContext(ctx).Where("wechat_open_id=?", openId).First(&user).Error
	return user, err
}
