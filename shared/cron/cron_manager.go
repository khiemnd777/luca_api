package cron

import (
	"fmt"
	"log"

	"github.com/khiemnd777/andy_api/shared/logger"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

type CronJob interface {
	Name() string
	DefaultSchedule() string
	Run() error
	ConfigKey() string // ex: "cron.cleanup_token"
}

var registeredJobs []CronJob

func RegisterJob(job CronJob) {
	registeredJobs = append(registeredJobs, job)
	log.Printf("[Cron job] %s is registering...", job.Name())
}

func StartAllCrons() {
	if len(registeredJobs) == 0 {
		logger.Info("No cron jobs registered")
		return
	}

	c := cron.New()
	for _, job := range registeredJobs {
		if !isJobEnabled(job) {
			logger.Info("Cron job disabled by config: " + job.Name())
			continue
		}

		schedule := getJobSchedule(job)
		log.Printf("Cron %s in schedule: %s", job.Name(), schedule)

		job := job
		_, err := c.AddFunc(schedule, func() {
			logger.Info("Running cron job: " + job.Name())
			if err := job.Run(); err != nil {
				logger.Error(fmt.Sprintf("Cron job failed: %s | Error: %s", job.Name(), err.Error()))
			}
		})
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to register cron job: %s | Error: %s", job.Name(), err.Error()))
		} else {
			logger.Info(fmt.Sprintf("Registered cron job: %s| Schedule: %s", job.Name(), schedule))
		}
	}
	c.Start()
	logger.Info(fmt.Sprintf("Cron scheduler started with %d jobs", len(registeredJobs)))
}

func isJobEnabled(job CronJob) bool {
	key := job.ConfigKey() + ".enabled"
	return viper.GetBool(key)
}

func getJobSchedule(job CronJob) string {
	key := job.ConfigKey() + ".schedule"
	schedule := viper.GetString(key)
	if schedule == "" {
		schedule = job.DefaultSchedule()
	}
	return schedule
}
