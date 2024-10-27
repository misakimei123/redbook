package prometheus

import (
	"context"
	"time"

	"github.com/misakimei123/redbook/internal/service/sms"
	"github.com/prometheus/client_golang/prometheus"
)

type SmsCounter struct {
	svc    sms.SMSService
	vector *prometheus.SummaryVec
}

func NewSmsCounter(svc sms.SMSService, opts prometheus.SummaryOpts) *SmsCounter {
	vec := prometheus.NewSummaryVec(opts, []string{"tpl_id"})
	prometheus.MustRegister(vec)
	return &SmsCounter{svc: svc, vector: vec}
}

func (s *SmsCounter) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	start := time.Now()
	defer func() {
		dur := time.Since(start).Milliseconds()
		s.vector.WithLabelValues(tplId).Observe(float64(dur))
	}()
	return s.svc.Send(ctx, tplId, args, number...)
}
