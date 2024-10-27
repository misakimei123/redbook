package failover

import (
	"errors"
	"log"
	"sync/atomic"

	"context"

	"github.com/misakimei123/redbook/internal/service/sms"
)

var errAllSMSFailed = errors.New("all sms failed")

type FailOverSMSService struct {
	svcs []sms.SMSService
	idx  uint64
}

func (s *FailOverSMSService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	for _, svc := range s.svcs {
		err := svc.Send(ctx, tplId, args, number...)
		if err == nil {
			return nil
		}
		log.Println(err)
	}
	return errAllSMSFailed
}

func (s *FailOverSMSService) SendV1(ctx context.Context, tplId string, args []string, number ...string) error {
	idx := atomic.AddUint64(&s.idx, 1)
	length := uint64(len(s.svcs))
	for i := idx; i < length+idx; i++ {
		svc := s.svcs[i%length]
		err := svc.Send(ctx, tplId, args, number...)
		switch err {
		case nil:
			return nil
		case context.Canceled, context.DeadlineExceeded:
			return err
		}
		log.Println(err)
	}
	return errAllSMSFailed
}

func NewFailOverSMSService(svcs ...sms.SMSService) sms.SMSService {
	length := len(svcs)
	failOverSMSService := FailOverSMSService{svcs: make([]sms.SMSService, length)}
	for i := 0; i < length; i++ {
		failOverSMSService.svcs[i] = svcs[i]
	}
	return &failOverSMSService
}
