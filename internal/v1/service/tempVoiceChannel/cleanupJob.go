package tempVoiceChannel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/repo"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type CleanupJob struct {
	stateRepo *repo.TempVoiceChannelStateRepo
	session   *discordgo.Session
	log       *zap.SugaredLogger
}

type CleanupJobParams struct {
	fx.In
	StateRepo *repo.TempVoiceChannelStateRepo
	Session   *discordgo.Session
	Log       *zap.Logger
}

func NewCleanupJob(p CleanupJobParams) *CleanupJob {
	return &CleanupJob{
		stateRepo: p.StateRepo,
		session:   p.Session,
		log:       p.Log.Sugar(),
	}
}

func (j *CleanupJob) Run(ctx context.Context) error {
	states, err := j.stateRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch active channels: %w", err)
	}

	for _, state := range states {
		channel, err := j.session.Channel(state.ChannelID, discordgo.WithContext(ctx))
		if err != nil {
			var restErr *discordgo.RESTError
			if ok := errors.As(err, &restErr); ok && restErr.Response != nil && restErr.Response.StatusCode == http.StatusNotFound {
				if err := j.stateRepo.DeleteByChannelID(ctx, state.ChannelID); err != nil {
					j.log.Warnw("failed to delete orphaned record", "channelID", state.ChannelID, zap.Error(err))
				}
			}
			continue
		}

		memberCount := countChannelMembers(j.session, channel.GuildID, state.ChannelID)
		if memberCount == 0 {
			if _, err := j.session.ChannelDelete(state.ChannelID, discordgo.WithContext(ctx)); err != nil {
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
