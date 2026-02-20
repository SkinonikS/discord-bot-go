package tempVoiceChannel

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Job struct {
	log *zap.Logger
}

type JobParams struct {
	fx.In
	Log *zap.Logger
}

func NewJob(p JobParams) *Job {
	return &Job{
		log: p.Log,
	}
}

func (j *Job) Run() error {
	return nil
}
