package job

import (
	"strconv"
	"time"

	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
)

type cronJobAdapterFunc func()

type CronJobBuilder struct {
	l      logger.LoggerV1
	vector *prometheus.SummaryVec
}

func (c cronJobAdapterFunc) Run() {
	c()
}

func NewJobBuilder(l logger.LoggerV1, opt prometheus.SummaryOpts) *CronJobBuilder {
	vec := prometheus.NewSummaryVec(opt, []string{"job", "success"})
	prometheus.MustRegister(vec)
	return &CronJobBuilder{l: l, vector: vec}
}

func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobAdapterFunc(func() {
		now := time.Now()
		b.l.Debug("start job", logger.String("name", name))
		err := job.Run()
		if err != nil {
			b.l.Error("execute job fail", logger.Error(err), logger.String("name", name))
		}
		duration := time.Since(now)
		b.vector.WithLabelValues(name, strconv.FormatBool(err == nil)).
			Observe(float64(duration.Milliseconds()))
	})
}
