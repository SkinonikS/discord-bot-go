package reactionRole

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/v1/database/model"
	"github.com/SkinonikS/discord-bot-go/internal/v1/database/repo"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type CleanupJob struct {
	repo    *repo.ReactionRoleRepo
	session *discordgo.Session
	log     *zap.SugaredLogger
}

type CleanupJobParams struct {
	fx.In
	Repo    *repo.ReactionRoleRepo
	Session *discordgo.Session
	Log     *zap.Logger
}

func NewCleanupJob(p CleanupJobParams) *CleanupJob {
	return &CleanupJob{
		repo:    p.Repo,
		session: p.Session,
		log:     p.Log.Sugar(),
	}
}

func (j *CleanupJob) Run(ctx context.Context) error {
	records, err := j.repo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch reaction roles: %w", err)
	}

	type msgKey struct{ channelID, messageID string }
	groups := make(map[msgKey][]*model.ReactionRole, len(records))
	for _, r := range records {
		key := msgKey{r.ChannelID, r.MessageID}
		groups[key] = append(groups[key], r)
	}

	for key, roles := range groups {
		msg, err := j.session.ChannelMessage(key.channelID, key.messageID, discordgo.WithContext(ctx))
		if err != nil {
			var restErr *discordgo.RESTError
			if errors.As(err, &restErr) && restErr.Response != nil && restErr.Response.StatusCode == http.StatusNotFound {
				if err := j.repo.DeleteByMessageID(ctx, key.messageID); err != nil {
					j.log.Warnw("failed to delete records for missing message", "messageID", key.messageID, zap.Error(err))
				}
			}
			continue
		}

		botReacted := make(map[string]bool, len(msg.Reactions))
		for _, r := range msg.Reactions {
			if r.Me {
				botReacted[emojiKey(*r.Emoji)] = true
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
