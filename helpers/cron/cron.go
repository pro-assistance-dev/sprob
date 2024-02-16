package cron

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// CustomCron struct
type Cron struct {
	cron *cron.Cron
}

type Job struct {
	Schedule string
	Function func()
}

type Jobs []*Job

// NewCronHelper func
func NewCron() (cr *Cron) {
	location := time.FixedZone("UTC+3", 3*60*60)
	return &Cron{cron: cron.New(cron.WithLocation(location))}
}

func (cr Cron) AddJobs(jobs Jobs) (err error) {
	for _, job := range jobs {
		_, err = cr.cron.AddFunc(job.Schedule, job.Function)
		if err != nil {
			log.Fatal(job)
		}
	}
	return
}
