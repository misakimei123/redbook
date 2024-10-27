package dao

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

type ProfileDao interface {
	Update(ctx context.Context, profile *Profile) error
	FindByUserId(ctx context.Context, userId int64) (Profile, error)
}

type GormProfileDao struct {
	db *gorm.DB
}

func (d *GormProfileDao) Update(ctx context.Context, profile *Profile) error {
	if d.db.WithContext(ctx).Model(profile).Where(Profile{UserId: profile.UserId}).Updates(profile).RowsAffected == 0 {
		return d.db.Create(profile).Error
	}
	return nil
}

func (d *GormProfileDao) FindByUserId(ctx context.Context, userId int64) (Profile, error) {
	var profile Profile
	err := d.db.WithContext(ctx).Where(Profile{UserId: userId}).First(&profile).Error
	return profile, err
}

func NewGormProfileDao(db *gorm.DB) ProfileDao {
	return &GormProfileDao{
		db,
	}
}

type Profile struct {
	User     User  `gorm:"association_foreignkey:Id"`
	UserId   int64 `gorm:"unique"`
	Birthday sql.NullInt64
	Nick     sql.NullString
	AboutMe  sql.NullString
}
