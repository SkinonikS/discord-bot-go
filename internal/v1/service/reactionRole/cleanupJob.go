package reactionRole

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

	DiscordConfig *discord.Config
	Service       Service
	Log           *zap.Logger
}

func NewCleanupJob(p CleanupJobParams) *CleanupJob {
	return &CleanupJob{
		discordConfig: p.DiscordConfig,
		service:       p.Service,
		log:           p.Log.Sugar(),
	}
}

func (j *CleanupJob) Task() gocron.Task {
	return gocron.NewTask(func(ctx context.Context) {
		if err := j.service.DeleteStaleReactions(ctx, j.discordConfig.ShardID); err != nil {
			j.log.Errorw("failed to cleanup orphaned reaction roles", zap.Error(err))
		} else {
			j.log.Infow("cleaned up orphaned reaction roles")
		}
	})
}

func (j *CleanupJob) Definition() gocron.JobDefinition {
	return gocron.DurationJob(10 * time.Minute)
}
