package tempVoiceChannel

import (
	"context"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	"github.com/go-co-op/gocron/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type CleanupJob struct {
	service       Service
	discordConfig *discord.Config
	log           *zap.SugaredLogger
}

type CleanupJobParams struct {
	fx.In

	Service       Service
	DiscordConfig *discord.Config
	Log           *zap.Logger
}

func NewCleanupJob(p CleanupJobParams) *CleanupJob {
	return &CleanupJob{
		service:       p.Service,
		discordConfig: p.DiscordConfig,
		log:           p.Log.Sugar(),
	}
}

func (j *CleanupJob) Task() gocron.Task {
	return gocron.NewTask(func(ctx context.Context) {
		if err := j.service.DeleteStaleChannels(ctx, j.discordConfig.ShardID); err != nil {
			j.log.Warnw("failed to delete stale channels", zap.Error(err))
		} else {
			j.log.Infow("cleaned up stale channels")
		}
	})
}

func (j *CleanupJob) Definition() gocron.JobDefinition {
	return gocron.DurationJob(10 * time.Minute)
}
