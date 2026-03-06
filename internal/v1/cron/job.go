package cron

import "github.com/go-co-op/gocron/v2"

type Job interface {
	Definition() gocron.JobDefinition
	Task() gocron.Task
}
