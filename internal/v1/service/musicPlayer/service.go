package musicPlayer

import (
	"context"
	"sync"
	"time"

	disgobot "github.com/disgoorg/disgo/bot"
	disgolavalink "github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
)

type Service interface {
	StartTimer(st StartTimer) *time.Timer
	StopTimer(guildID snowflake.ID) bool
}

type serviceImpl struct {
	botClient    *disgobot.Client
	idleTimers   map[snowflake.ID]*time.Timer
	idleTimersMu sync.Mutex
	config       *Config
}

type ServiceParams struct {
	fx.In

	Config    *Config
	BotClient *disgobot.Client
}

func NewService(p ServiceParams) Service {
	return &serviceImpl{
		botClient:  p.BotClient,
		config:     p.Config,
		idleTimers: make(map[snowflake.ID]*time.Timer),
	}
}

type StartTimer struct {
	GuildID snowflake.ID
	Track   func() *disgolavalink.Track
}

func (s *serviceImpl) StartTimer(st StartTimer) *time.Timer {
	s.idleTimersMu.Lock()
	defer s.idleTimersMu.Unlock()

	if t, ok := s.idleTimers[st.GuildID]; ok {
		t.Stop()
	}

	s.idleTimers[st.GuildID] = time.AfterFunc(s.config.IdleTimeout, func() {
		s.idleTimersMu.Lock()
		delete(s.idleTimers, st.GuildID)
		s.idleTimersMu.Unlock()

		if st.Track() != nil {
			return
		}

		_ = s.botClient.UpdateVoiceState(context.Background(), st.GuildID, nil, false, false)
	})

	return s.idleTimers[st.GuildID]
}

func (s *serviceImpl) StopTimer(guildID snowflake.ID) bool {
	s.idleTimersMu.Lock()
	defer s.idleTimersMu.Unlock()
	if t, ok := s.idleTimers[guildID]; ok {
		t.Stop()
		delete(s.idleTimers, guildID)

		return true
	}

	return false
}
