package tempVoiceChannel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/repo"
	disgobot "github.com/disgoorg/disgo/bot"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/go-co-op/gocron/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type CleanupJob struct {
	stateRepo *repo.TempVoiceChannelStateRepo
	client    *disgobot.Client
	log       *zap.SugaredLogger
}

type CleanupJobParams struct {
	fx.In
	StateRepo *repo.TempVoiceChannelStateRepo
	Client    *disgobot.Client
	Log       *zap.Logger
}

func NewCleanupJob(p CleanupJobParams) *CleanupJob {
	return &CleanupJob{
		stateRepo: p.StateRepo,
		client:    p.Client,
		log:       p.Log.Sugar(),
	}
}

func (j *CleanupJob) Task() gocron.Task {
	return gocron.NewTask(j.run)
}

func (j *CleanupJob) Definition() gocron.JobDefinition {
	return gocron.DurationJob(5 * time.Minute)
}

func (j *CleanupJob) run(ctx context.Context) error {
	states, err := j.stateRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch active channels: %w", err)
	}

	for _, state := range states {
		channel, err := j.client.Rest.GetChannel(state.ChannelID, disgorest.WithCtx(ctx))
		if err != nil {
			if restErr, ok := errors.AsType[*disgorest.Error](err); ok && restErr.Response != nil && restErr.Response.StatusCode == http.StatusNotFound {
				if err := j.stateRepo.DeleteByChannelID(ctx, state.ChannelID); err != nil {
					j.log.Warnw("failed to delete orphaned record", "channelID", state.ChannelID, zap.Error(err))
				}
			}
			continue
		}

		memberCount := countChannelMembers(j.client, channel.ID(), state.ChannelID)
		if memberCount == 0 {
			if err := j.client.Rest.DeleteChannel(state.ChannelID, disgorest.WithCtx(ctx)); err != nil {
				j.log.Warnw("failed to delete empty channel", "channelID", state.ChannelID, zap.Error(err))
				continue
			}
			if err := j.stateRepo.DeleteByChannelID(ctx, state.ChannelID); err != nil {
				j.log.Warnw("failed to delete state record", "channelID", state.ChannelID, zap.Error(err))
			}
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}
