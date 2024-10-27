package dao

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SMS struct {
	Id     int64 `gorm:"primaryKey, autoIncrement"`
	Paras  string
	Status string
}

var ErrNoSMS = errors.New("no buffered sms need to send now")

type SMSDao interface {
	Insert(ctx context.Context, sms SMS) error
	QueryAndUpdate(ctx context.Context, oldStatus string, newStatus string) (SMS, error)
	Update(ctx context.Context, id int64, status string) error
}

type SMSGormDao struct {
	db *gorm.DB
}

func (s *SMSGormDao) Insert(ctx context.Context, sms SMS) error {
	return s.db.WithContext(ctx).Create(sms).Error
}

func (s *SMSGormDao) QueryAndUpdate(ctx context.Context, oldStatus string, newStatus string) (SMS, error) {
	var sms SMS
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 使用悲观锁，锁住记录
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("status = ?", oldStatus).First(&sms)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return ErrNoSMS
		}

		sms.Status = newStatus
		err := tx.Save(&sms).Error
		return err
	})
	return sms, err
}

func (s *SMSGormDao) Update(ctx context.Context, id int64, status string) error {
	var sms SMS
	result := s.db.WithContext(ctx).First(&sms, id)
	if result.Error != nil {
		return result.Error
	}
	sms.Status = status
	result = s.db.WithContext(ctx).Save(&sms)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func NewSMSGormDao(db *gorm.DB) SMSDao {
	return &SMSGormDao{
		db: db,
	}
}
