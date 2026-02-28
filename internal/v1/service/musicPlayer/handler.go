package musicPlayer

import (
	"github.com/SkinonikS/discord-bot-go/internal/v1/util"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type Handler struct {
	manager *Manager
	log     *zap.SugaredLogger
}

func NewHandler(manager *Manager, log *zap.Logger) *Handler {
	return &Handler{
		manager: manager,
		log:     log.Sugar(),
	}
}

func (h *Handler) Handle(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	err := util.Safe(func() error {
		if s.State.User == nil || e.UserID != s.State.User.ID {
			return nil
		}

		if e.ChannelID == "" {
			return h.manager.Disconnect(e.GuildID)
		}

		if !h.manager.HasGuildState(e.GuildID) {
			s.RLock()
			vc := s.VoiceConnections[e.GuildID]
			s.RUnlock()
			if vc != nil {
				_ = vc.Disconnect()
			}
		}

		return nil
	})
	if err != nil {
		h.log.Errorw("failed to handle voice state update", zap.Error(err))
	}
}
