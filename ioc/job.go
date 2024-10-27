package ioc

import (
	"github.com/misakimei123/redbook/internal/job"
	"github.com/misakimei123/redbook/internal/service"
	"github.com/misakimei123/redbook/pkg/distribute/balance"
	"github.com/misakimei123/redbook/pkg/distribute/lock"
	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
)

func InitRankingJob(svc service.RankingService, l logger.LoggerV1, dLock lock.Lock,
	balancer balance.LoadBalance[job.RunNode]) job.Job {
	return job.NewRankingJob(svc, l, dLock, balancer)
}

func InitJobs(runJob job.Job, l logger.LoggerV1) *cron.Cron {
	builder := job.NewJobBuilder(l, prometheus.SummaryOpts{
		Namespace: "misakimei123",
		Subsystem: "redbook",
		Name:      "cronjob",
		Objectives: map[float64]float64{
			0.5:   0.1,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	c := cron.New(cron.WithSeconds())
	_, err := c.AddJob("@every 30s", builder.Build(runJob))
	if err != nil {
		panic(err)
	}
	return c
}
