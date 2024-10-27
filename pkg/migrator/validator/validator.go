package validator

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/misakimei123/redbook/pkg/migrator"
	"github.com/misakimei123/redbook/pkg/migrator/events"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type Validator[T migrator.Entity] struct {
	base          *gorm.DB
	target        *gorm.DB
	l             logger.LoggerV1
	producer      events.Producer
	direction     string
	batchSize     int
	utime         int64
	sleepInterval time.Duration
	fromBase      func(ctx context.Context, offset int) ([]T, error)
}

func NewValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB, l logger.LoggerV1, producer events.Producer, direction string) *Validator[T] {
	v := &Validator[T]{base: base, target: target, l: l, producer: producer, direction: direction}
	return v.Full()
}

func (v *Validator[T]) Utime(utime int64) *Validator[T] {
	v.utime = utime
	return v
}

func (v *Validator[T]) SleepInterval(i time.Duration) *Validator[T] {
	v.sleepInterval = i
	return v
}

func (v *Validator[T]) Full() *Validator[T] {
	v.fromBase = v.fullFromBase
	return v
}

func (v *Validator[T]) Incr() *Validator[T] {
	v.fromBase = v.incrFromBase
	return v
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		return v.ValidateBase2Target(ctx)
	})
	eg.Go(func() error {
		return v.ValidateTarget2Base(ctx)
	})
	return eg.Wait()
}

func (v *Validator[T]) fullFromBase(ctx context.Context, offset int) ([]T, error) {
	dbCtx, cancelFunc := context.WithTimeout(ctx, time.Second)
	defer cancelFunc()
	var src []T
	err := v.base.WithContext(dbCtx).Order("id").Offset(offset).Limit(v.batchSize).Find(&src).Error
	return src, err
}

func (v *Validator[T]) incrFromBase(ctx context.Context, offset int) ([]T, error) {
	dbCtx, cancelFunc := context.WithTimeout(ctx, time.Second)
	defer cancelFunc()
	var src []T
	err := v.base.WithContext(dbCtx).Where("utime > ?", v.utime).Order("utime").Offset(offset).Limit(v.batchSize).Find(&src).Error
	return src, err
}

func (v *Validator[T]) ValidateBase2Target(ctx context.Context) error {
	offset := 0
	for {
		srcs, err := v.fromBase(ctx, offset)
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			v.l.Error("query base fail", logger.Error(err))
			offset += len(srcs)
			continue
		}
		ids := slice.Map(srcs, func(idx int, t T) int64 {
			return t.ID()
		})
		var dsts []T
		err = v.target.WithContext(ctx).Where("id in ?", ids).Find(&dsts).Error
		switch err {
		case gorm.ErrRecordNotFound:
			v.notifyFix(srcs, events.InconsistentEventTypeTargetMissing)

		case nil:
			//diff is in target not in base
			diff := slice.DiffSetFunc(srcs, dsts, func(src, dst T) bool {
				return src.CompareTo(dst)
			})
			if len(diff) > 0 {
				v.notifyFix(diff, events.InconsistentEventTypeNEQ)
			}
		default:
			v.l.Error("query target fail", logger.Error(err))
		}
		offset += len(srcs)
	}
}

func (v *Validator[T]) ValidateTarget2Base(ctx context.Context) error {
	offset := 0
	for {
		var ts []T
		err := v.target.WithContext(ctx).Order("id").Offset(offset).Limit(v.batchSize).Find(&ts).Error
		if err == gorm.ErrRecordNotFound || len(ts) == 0 {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}

		if err != nil {
			v.l.Error("target 2 base query target fail", logger.Error(err))
			offset += len(ts)
			continue
		}

		ids := slice.Map(ts, func(idx int, t T) int64 {
			return t.ID()
		})
		var srcTs []T
		err = v.base.WithContext(ctx).Select("id").Where("id in ?", ids).Find(&srcTs).Error
		if err == gorm.ErrRecordNotFound || len(srcTs) == 0 {
			v.notifyFix(ts, events.InconsistentEventTypeBaseMissing)
			offset += len(ts)
			continue
		}
		if err != nil {
			v.l.Error("target 2 base query src fail", logger.Error(err))
			offset += len(ts)
			continue
		}
		//diff is in target not in base
		diff := slice.DiffSetFunc(ts, srcTs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.notifyFix(diff, events.InconsistentEventTypeBaseMissing)
		if len(srcTs) < v.batchSize {
			if v.sleepInterval <= 0 {
				return nil
			}
			offset += len(ts)
			time.Sleep(v.sleepInterval)
		}
		offset += len(ts)
	}
}

func (v *Validator[T]) notifyFix(ts []T, evt string) {
	for _, t := range ts {
		v.notify(t.ID(), evt)
	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()
	err := v.producer.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
		ID:        id,
		Direction: v.direction,
		Type:      typ,
	})
	if err != nil {
		v.l.Error("send message fail",
			logger.Int64("id", id),
			logger.String("type", typ),
			logger.String("direction", v.direction))
	}
}
