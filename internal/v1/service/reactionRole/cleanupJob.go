package reactionRole

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/model"
	"github.com/SkinonikS/discord-bot-go/internal/v1/database/repo"
	disgobot "github.com/disgoorg/disgo/bot"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/go-co-op/gocron/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type CleanupJob struct {
	repo   *repo.ReactionRoleRepo
	client *disgobot.Client
	log    *zap.SugaredLogger
}

type CleanupJobParams struct {
	fx.In
	Repo   *repo.ReactionRoleRepo
	Client *disgobot.Client
	Log    *zap.Logger
}

func NewCleanupJob(p CleanupJobParams) *CleanupJob {
	return &CleanupJob{
		repo:   p.Repo,
		client: p.Client,
		log:    p.Log.Sugar(),
	}
}

func (j *CleanupJob) Task() gocron.Task {
	return gocron.NewTask(j.run)
}

func (j *CleanupJob) Definition() gocron.JobDefinition {
	return gocron.DurationJob(5 * time.Minute)
}

func (j *CleanupJob) run(ctx context.Context) error {
	records, err := j.repo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch reaction roles: %w", err)
	}

	type msgKey struct{ channelID, messageID snowflake.ID }
	groups := make(map[msgKey][]*model.ReactionRole, len(records))
	for _, r := range records {
		key := msgKey{r.ChannelID, r.MessageID}
		groups[key] = append(groups[key], r)
	}

	for key, roles := range groups {
		msg, err := j.client.Rest.GetMessage(key.channelID, key.messageID, disgorest.WithCtx(ctx))
		if err != nil {
			if restErr, ok := errors.AsType[*disgorest.Error](err); ok && restErr.Response != nil && restErr.Response.StatusCode == http.StatusNotFound {
				if err := j.repo.DeleteByMessageID(ctx, key.messageID); err != nil {
					j.log.Warnw("failed to delete records for missing message", "messageID", key.messageID, zap.Error(err))
				}
			}
			continue
		}

		botReacted := make(map[string]bool, len(msg.Reactions))
		for _, r := range msg.Reactions {
			if r.Me {
				botReacted[r.Emoji.Reaction()] = true
			}
		}

		for _, role := range roles {
			if !botReacted[role.EmojiName] {
				if err := j.repo.DeleteByMessageAndEmoji(ctx, role.MessageID, role.EmojiName); err != nil {
					j.log.Warnw("failed to delete orphaned reaction role", "messageID", role.MessageID, "emoji", role.EmojiName, zap.Error(err))
				}
			}
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}
