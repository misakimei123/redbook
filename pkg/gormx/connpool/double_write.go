package connpool

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/misakimei123/redbook/pkg/logger"
	"gorm.io/gorm"
)

const (
	PatternSrcOnly  = "src_only"
	PatternSrcFirst = "src_first"
	PatternDstFirst = "dst_first"
	PatternDstOnly  = "dst_only"
)

var errUnknownPattern = errors.New("unknown double write pattern")

type DoubleWritePool struct {
	src     gorm.ConnPool
	dst     gorm.ConnPool
	pattern *atomicx.Value[string]
	l       logger.LoggerV1
}

func NewDoubleWritePool(src *gorm.DB, dst *gorm.DB, l logger.LoggerV1) *DoubleWritePool {
	return &DoubleWritePool{src: src.ConnPool, dst: dst.ConnPool, l: l, pattern: atomicx.NewValueOf(PatternSrcOnly)}
}

func (d *DoubleWritePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWriteTx{src: src, l: d.l, pattern: pattern}, err
	case PatternSrcFirst:
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			d.l.Error("double write PatternSrcFirst dst fail", logger.Error(err))
		}
		return &DoubleWriteTx{src: src, dst: dst, l: d.l, pattern: pattern}, nil
	case PatternDstOnly:
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWriteTx{dst: dst}, err
	case PatternDstFirst:
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			d.l.Error("double write PatternDstFirst src fail", logger.Error(err))
		}
		return &DoubleWriteTx{src: src, dst: dst, l: d.l, pattern: pattern}, nil
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	//TODO implement me
	panic("not support in write pattern")
}

func (d *DoubleWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		result, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		_, err = d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			d.l.Error("double write dst fail", logger.Error(err))
		}
		return result, nil
	case PatternDstFirst:
		return d.src.ExecContext(ctx, query, args...)
	case PatternDstOnly:
		result, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		_, err = d.src.ExecContext(ctx, query, args...)
		if err != nil {
			d.l.Error("double write src fail", logger.Error(err))
		}
		return result, nil
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		panic(errUnknownPattern)
	}
}

func (d *DoubleWritePool) UpdatePattern(pattern string) error {
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst, PatternDstFirst, PatternDstOnly:
		d.pattern.Store(pattern)
	default:
		return errUnknownPattern
	}
	return nil
}

type DoubleWriteTx struct {
	src     *sql.Tx
	dst     *sql.Tx
	pattern string
	l       logger.LoggerV1
}

func (d *DoubleWriteTx) Commit() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Commit()
	case PatternSrcFirst:
		err := d.src.Commit()
		if err != nil {
			return err
		}
		if d.dst == nil {
			return nil
		}
		err = d.dst.Commit()
		if err != nil {
			d.l.Error("target table commit fail", logger.Error(err))
		}
	case PatternDstOnly:
		return d.dst.Commit()
	case PatternDstFirst:
		err := d.dst.Commit()
		if err != nil {
			return err
		}
		if d.src == nil {
			return nil
		}
		err = d.src.Commit()
		if err != nil {
			d.l.Error("src table commit fail", logger.Error(err))
		}
	default:
		return errUnknownPattern
	}
	return nil
}

func (d *DoubleWriteTx) Rollback() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Rollback()
	case PatternSrcFirst:
		err := d.src.Rollback()
		if err != nil {
			return err
		}
		if d.dst == nil {
			return nil
		}
		err = d.dst.Rollback()
		if err != nil {
			d.l.Error("target table commit fail", logger.Error(err))
		}
	case PatternDstOnly:
		return d.dst.Rollback()
	case PatternDstFirst:
		err := d.dst.Rollback()
		if err != nil {
			return err
		}
		if d.src == nil {
			return nil
		}
		err = d.src.Rollback()
		if err != nil {
			d.l.Error("src table commit fail", logger.Error(err))
		}
	default:
		return errUnknownPattern
	}
	return nil
}

func (d *DoubleWriteTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	//TODO implement me
	panic("not support in write pattern")
}

func (d *DoubleWriteTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		result, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		if d.dst == nil {
			return result, nil
		}
		_, err = d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			d.l.Error("double write dst fail", logger.Error(err))
		}
		return result, nil
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	case PatternDstFirst:
		result, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		if d.src == nil {
			return result, nil
		}
		_, err = d.src.ExecContext(ctx, query, args...)
		if err != nil {
			d.l.Error("double write src fail", logger.Error(err))
		}
		return result, nil
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWriteTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWriteTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		panic(errUnknownPattern)
	}
}
