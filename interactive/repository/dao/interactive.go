package dao

import (
	"context"
	"time"

	"github.com/misakimei123/redbook/pkg/migrator"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InteractiveDao interface {
	IncrReadCnt(ctx context.Context, bizStr string, bizId int64) error
	InsertLikeInfo(ctx context.Context, bizStr string, bizId int64, uid int64) error
	DeleteLikeInfo(ctx context.Context, bizStr string, bizId int64, uid int64) error
	InsertCollectBiz(ctx context.Context, userCollectBiz UserCollectBiz) error
	GetLikeInfo(ctx context.Context, bizStr string, bizId int64, uid int64) (UserLikeBiz, error)
	GetCollectInfo(ctx context.Context, bizStr string, bizId int64, uid int64) (UserCollectBiz, error)
	Get(ctx context.Context, bizStr string, bizId int64) (Interactive, error)
	BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error
	GetByIds(ctx context.Context, bizStr string, ids []int64) ([]Interactive, error)
}

func NewInteractiveGormDao(db *gorm.DB) InteractiveDao {
	return &InteractiveGormDao{db: db}
}

type InteractiveGormDao struct {
	db *gorm.DB
}

func (i *InteractiveGormDao) GetByIds(ctx context.Context, bizStr string, ids []int64) ([]Interactive, error) {
	var res []Interactive
	err := i.db.WithContext(ctx).Where("biz_str = ? AND biz_id IN ?", bizStr, ids).Find(&res).Error
	return res, err
}

func (i *InteractiveGormDao) Get(ctx context.Context, bizStr string, bizId int64) (Interactive, error) {
	var intra Interactive
	err := i.db.WithContext(ctx).
		Where("biz_id=? and biz_str=?", bizId, bizStr).First(&intra).Error
	return intra, err
}

func (i *InteractiveGormDao) GetLikeInfo(ctx context.Context, bizStr string, bizId int64, uid int64) (UserLikeBiz, error) {
	var like UserLikeBiz
	err := i.db.WithContext(ctx).
		Where("uid=? and biz_id=? and biz_str=? and status=?", uid, bizId, bizStr, 1).First(&like).Error
	return like, err
}

func (i *InteractiveGormDao) GetCollectInfo(ctx context.Context, bizStr string, bizId int64, uid int64) (UserCollectBiz, error) {
	var collect UserCollectBiz
	err := i.db.WithContext(ctx).
		Where("uid=? and biz_id=? and biz_str=?", uid, bizId, bizStr).First(&collect).Error
	return collect, err
}

func (i *InteractiveGormDao) InsertCollectBiz(ctx context.Context, userCollectBiz UserCollectBiz) error {
	now := time.Now().UnixMilli()
	userCollectBiz.Ctime = now
	userCollectBiz.Utime = now
	return i.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&userCollectBiz).Error
		if err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"collect_cnt": gorm.Expr("`collect_cnt` + 1"),
				"utime":       now,
			}),
			UpdateAll: false,
		}).Create(&Interactive{
			BizId:      userCollectBiz.BizId,
			BizStr:     userCollectBiz.BizStr,
			CollectCnt: 1,
			Utime:      now,
			Ctime:      now,
		}).Error
	})
}

func (i *InteractiveGormDao) DeleteLikeInfo(ctx context.Context, bizStr string, bizId int64, uid int64) error {
	now := time.Now().UnixMilli()
	return i.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(UserLikeBiz{}).
			Where("uid=? and biz_id=? and biz_str=?", uid, bizId, bizStr).
			Updates(map[string]interface{}{
				"utime":  now,
				"status": 0,
			}).Error
		if err != nil {
			return err
		}
		return tx.Model(&Interactive{}).
			Where("biz_id=? and biz_str=?", bizId, bizStr).
			Updates(map[string]interface{}{
				"like_cnt": gorm.Expr("`like_cnt` - 1"),
				"utime":    now,
			}).Error
	})
}

func (i *InteractiveGormDao) InsertLikeInfo(ctx context.Context, bizStr string, bizId int64, uid int64) error {
	now := time.Now().UnixMilli()
	return i.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"utime":  now,
				"status": 1,
			}),
		}).Create(&UserLikeBiz{
			Id:     0,
			Uid:    uid,
			BizId:  bizId,
			BizStr: bizStr,
			Status: 1,
			Utime:  now,
			Ctime:  now,
		}).Error
		if err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt": gorm.Expr("`like_cnt` + 1"),
				"utime":    now,
			}),
			UpdateAll: false,
		}).Create(&Interactive{
			BizId:   bizId,
			BizStr:  bizStr,
			LikeCnt: 1,
			Utime:   now,
			Ctime:   now,
		}).Error
	})
}

func (i *InteractiveGormDao) IncrReadCnt(ctx context.Context, bizStr string, bizId int64) error {
	now := time.Now().UnixMilli()
	return i.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"read_cnt": gorm.Expr("`read_cnt` + 1"),
			"utime":    now,
		}),
		UpdateAll: false,
	}).Create(&Interactive{
		BizId:   bizId,
		BizStr:  bizStr,
		ReadCnt: 1,
		Utime:   now,
		Ctime:   now,
	}).Error
}

func (i *InteractiveGormDao) BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error {
	return i.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDAO := NewInteractiveGormDao(tx)
		var err error
		for i, id := range ids {
			err = txDAO.IncrReadCnt(ctx, bizs[i], id)
		}
		return err
	})
}

type UserLikeBiz struct {
	Id     int64  `gorm:"primaryKey, autoIncrement"`
	Uid    int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId  int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizStr string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	Status int
	Utime  int64
	Ctime  int64
}

type UserCollectBiz struct {
	Id     int64  `gorm:"primaryKey, autoIncrement"`
	Uid    int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId  int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizStr string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	Cid    int64  `gorm:"index"`
	Utime  int64
	Ctime  int64
}

type Interactive struct {
	Id         int64  `gorm:"primaryKey, autoIncrement"`
	BizId      int64  `gorm:"uniqueIndex:biz_type_id"`
	BizStr     string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Utime      int64
	Ctime      int64
}

func (i Interactive) ID() int64 {
	return i.Id
}

func (i Interactive) CompareTo(dst migrator.Entity) bool {
	val, ok := dst.(Interactive)
	if !ok {
		return false
	}
	return i == val
}
