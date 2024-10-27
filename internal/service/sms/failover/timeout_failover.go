package failover

import (
	"context"
	"sync/atomic"

	"github.com/misakimei123/redbook/internal/service/sms"
)

type TimeoutFailOverSMSService struct {
	svcs      []sms.SMSService
	idx       int32
	cnt       int32
	threshold int32
}

func (t *TimeoutFailOverSMSService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)

	if cnt >= t.threshold {
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			atomic.StoreInt32(&t.cnt, 0)
		}
		idx = newIdx
	}

	err := t.svcs[idx].Send(ctx, tplId, args, number...)
	switch err {
	case nil:
		//非连续超时就清零
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1)
	default:
		//非超时错误，不增加计数，EOF可以考虑直接切
	}
	return err
}
