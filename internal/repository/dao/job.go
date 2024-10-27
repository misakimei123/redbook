package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Job struct {
	Id         int64 `gorm:"primaryKey, autoIncrement"`
	Expression string
	Executor   string
	Name       string `gorm:"type:varchar(128);unique"`
	Cfg        string
	Status     uint8
	NextTime   int64 `gorm:"index"`
	Version    int
	Ctime      int64
	Utime      int64
}

type GormJobDao struct {
	db         *gorm.DB
	jobTimeOut time.Duration
}

func NewGormJobDao(db *gorm.DB) JobDao {
	return &GormJobDao{db: db, jobTimeOut: time.Minute * 2}
}

const (
	jobStatusWaiting = iota
	jobStatusRunning
	jobStatusPaused
)

type JobDao interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, jid int64) error
	UpdateNextTime(ctx context.Context, jid int64, nextTime time.Time) error
}

func (g *GormJobDao) Preempt(ctx context.Context) (Job, error) {
	db := g.db
	for {
		var j Job
		now := time.Now().UnixMilli()
		err := db.WithContext(ctx).Where("(status = ? and next_time < ?) or (status = ? and utime <= ?)",
			jobStatusWaiting, now, jobStatusRunning, now-g.jobTimeOut.Milliseconds()).First(&j).Error
		if err != nil {
			return j, err
		}
		res := db.WithContext(ctx).Model(&Job{}).Where("id = ? and version = ?", j.Id, j.Version).Updates(map[string]any{
			"status":  jobStatusRunning,
			"utime":   now,
			"version": j.Version + 1,
		})
		if res.Error != nil {
			return j, res.Error
		}
		if res.RowsAffected == 0 {
			continue
		}
		return j, err
	}
}

func (g *GormJobDao) Release(ctx context.Context, jid int64) error {
	return g.db.WithContext(ctx).Model(&Job{}).Where("id = ?", jid).Updates(map[string]any{
		"status": jobStatusWaiting,
		"utime":  time.Now().UnixMilli(),
	}).Error
}

func (g *GormJobDao) UpdateUtime(ctx context.Context, jid int64) error {
	return g.db.WithContext(ctx).Model(&Job{}).Where("id = ?", jid).Updates(map[string]any{
		"utime": time.Now().UnixMilli(),
	}).Error
}

func (g *GormJobDao) UpdateNextTime(ctx context.Context, jid int64, nextTime time.Time) error {
	return g.db.WithContext(ctx).Model(&Job{}).Where("id = ?", jid).Updates(map[string]any{
		"next_time": nextTime.UnixMilli(),
	}).Error
}
