package async

import (
	"sync/atomic"
	"time"

	"context"

	"github.com/misakimei123/redbook/internal/service/sms"
	"github.com/misakimei123/redbook/internal/service/sms/repo"
	"github.com/misakimei123/redbook/pkg/limiter"
)

type AsyncSMS struct {
	svc     sms.SMSService
	limiter limiter.Limiter
	key     string
	repo    repo.SMSRepo

	//连续N个请求超阈值认为服务挂了
	cnt          int32
	crashed      atomic.Value
	sending      atomic.Value
	durThreshold int64
	durN         int32

	//间隔时间后启动任务发送异步消息，如果缓存的消息能发完则认为状态恢复了。
	interval int64
	retryN   int
}

type AsyncSMSOptions struct {
	key          string
	durThreshold int64
	durN         int32
	interval     int64
	retryN       int
}

func NewAsyncSMS(svc sms.SMSService, l limiter.Limiter, smsRepo repo.SMSRepo) *AsyncSMS {
	return NewAsyncSMSWithOptions(svc, l, smsRepo, AsyncSMSOptions{
		key:          "sms-limiter",
		durThreshold: 500,
		durN:         5,
		interval:     1000,
		retryN:       3,
	})
}

func (a *AsyncSMS) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	cnt := atomic.LoadInt32(&a.cnt)
	crashed := a.crashed.Load().(bool)
	limit, err := a.limiter.Limit(ctx, a.key)
	if err != nil {
		return err
	}
	//限速或者连续N个超阈值，则认为崩溃
	if crashed || cnt >= a.durN || limit {
		a.crashed.Store(true)
		return a.repo.Put(ctx, repo.SMSPara{
			TplId:   tplId,
			Args:    args,
			Numbers: numbers,
		})
	}

	now := time.Now()
	err = a.svc.Send(ctx, tplId, args, numbers...)
	switch err {
	case nil:
		elapse := time.Since(now)
		if elapse.Milliseconds() >= a.durThreshold {
			atomic.AddInt32(&a.cnt, 1)
		} else {
			//如果正常发送了，则认为状态正常了
			a.crashed.Store(false)
			atomic.StoreInt32(&a.cnt, 0)
		}
		return nil
	default:
		return err
	}
}

func NewAsyncSMSWithOptions(svc sms.SMSService, l limiter.Limiter, smsRepo repo.SMSRepo, options AsyncSMSOptions) *AsyncSMS {
	asyncSms := &AsyncSMS{
		svc:          svc,
		limiter:      l,
		repo:         smsRepo,
		key:          options.key,
		durThreshold: options.durThreshold,
		durN:         options.durN,
		interval:     options.interval,
		retryN:       options.retryN,
	}
	asyncSms.crashed.Store(false)
	asyncSms.sending.Store(false)
	go func() {
		asyncSms.asyncSend()
	}()
	return asyncSms
}

func (a *AsyncSMS) asyncSend() {
	time.Sleep(time.Duration(a.interval) * time.Millisecond)
	for {
		//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		smsPara, err := a.repo.Get(context.Background())
		//cancel()

		switch err {
		case repo.ErrNoSMS:
			swapped := a.sending.CompareAndSwap(true, false)
			if swapped {
				a.crashed.CompareAndSwap(true, false)
				atomic.StoreInt32(&a.cnt, 0)
			}

		case nil:
			a.sending.Store(true)
			i := 0
			for ; i < a.retryN; i++ {
				//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				err := a.svc.Send(context.Background(), smsPara.TplId, smsPara.Args, smsPara.Numbers...)
				//defer cancel()
				if err == nil {
					_ = a.repo.Del4Success(context.Background(), smsPara.Id)
					break
				}
				time.Sleep(time.Duration(a.interval) * time.Millisecond)
			}
			if i == a.retryN {
				_ = a.repo.Del4Fail(context.Background(), smsPara.Id)
			}

		default:
			time.Sleep(time.Duration(a.interval) * time.Millisecond)
		}
	}
}
