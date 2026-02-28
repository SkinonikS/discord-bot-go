package musicPlayer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/musicPlayer/player"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/musicPlayerSource"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	ErrNotInitialized    = errors.New("guild player not initialized")
	ErrNotInVoiceChannel = errors.New("you are not in a voice channel")
	ErrNotPlaying        = errors.New("not currently playing")
	ErrNotPaused         = errors.New("not currently paused")
)

type guildState struct {
	mu              sync.Mutex
	guildID         string
	textChannelID   string
	player          *player.Player
	voiceConnection *discordgo.VoiceConnection
}

type PlayParams struct {
	GuildID       string
	UserID        string
	TextChannelID string
	URL           string
	SourceName    string
	PlayNow       bool
}

type Manager struct {
	mu          sync.RWMutex
	idleTimeout time.Duration
	session     *discordgo.Session
	log         *zap.SugaredLogger
	sources     *musicPlayerSource.Registry
	ffmpegPath  string
	guildStates map[string]*guildState
}

type ManagerParams struct {
	fx.In
	Config  *Config
	Path    *foundation.Path
	Session *discordgo.Session
	Sources *musicPlayerSource.Registry
	Log     *zap.Logger
}

func NewManager(p ManagerParams) (*Manager, error) {
	ffmpegPath, err := p.Path.ResolveFrom(p.Config.FfmpegPath)
	if err != nil {
		return nil, err
	}

	return &Manager{
		idleTimeout: p.Config.IdleTimeout,
		session:     p.Session,
		ffmpegPath:  ffmpegPath,
		sources:     p.Sources,
		log:         p.Log.Sugar(),
		guildStates: make(map[string]*guildState),
	}, nil
}

func (p *Manager) Play(ctx context.Context, playParams PlayParams) (*player.Track, error) {
	gs, err := p.iniGuildState(ctx, playParams.GuildID, playParams.GuildID, playParams.UserID)
	if err != nil {
		return nil, err
	}

	track, err := p.resolveTrack(ctx, playParams.UserID, playParams.URL, playParams.SourceName)
	if err != nil {
		return nil, err
	}

	if playParams.PlayNow {
		gs.player.ShiftQueue(track)
	} else {
		gs.player.PushQueue(track)
	}

	if gs.player.Status() == player.StatusIdle {
		go gs.player.Play(p.newTrackEncoder(gs, playParams.GuildID))
	} else {
		if playParams.PlayNow {
			gs.player.Skip()
		}
	}

	return track, nil
}

func (p *Manager) Skip(guildID string) error {
	return p.withGuildState(guildID, func(gs *guildState) error {
		gs.player.Skip()
		return nil
	})
}

func (p *Manager) Stop(guildID string) error {
	return p.withGuildState(guildID, func(gs *guildState) error {
		gs.player.Stop()
		return nil
	})
}

func (p *Manager) Remove(guildID string, index int) (*player.Track, error) {
	var track *player.Track
	err := p.withGuildState(guildID, func(gs *guildState) error {
		t, err := gs.player.Remove(index)
		if err != nil {
			return err
		}
		track = t
		return nil
	})
	return track, err
}

func (p *Manager) Pause(guildID string) error {
	return p.withGuildState(guildID, func(gs *guildState) error {
		if !gs.player.Pause() {
			return ErrNotPlaying
		}
		return nil
	})
}

func (p *Manager) Resume(guildID string) error {
	return p.withGuildState(guildID, func(gs *guildState) error {
		if !gs.player.Resume() {
			return ErrNotPaused
		}
		return nil
	})
}

func (p *Manager) Queue(guildID string) (*player.Queue, error) {
	var queue *player.Queue
	err := p.withGuildState(guildID, func(gs *guildState) error {
		queue = gs.player.Queue()
		current := gs.player.Current()
		if current != nil {
			queue.Shift(current)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return queue, nil
}

func (p *Manager) Disconnect(guildID string) error {
	p.mu.RLock()
	gs, ok := p.guildStates[guildID]
	p.mu.RUnlock()
	if !ok {
		return nil
	}

	gs.player.Stop()

	gs.mu.Lock()
	vc := gs.voiceConnection
	gs.mu.Unlock()
	if vc != nil {
		_ = vc.Disconnect()
	}

	p.removeGuildState(guildID)
	return nil
}

func (p *Manager) HasGuildState(guildID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.guildStates[guildID]
	return ok
}

func (p *Manager) newTrackEncoder(gs *guildState, guildID string) *TrackEncoder {
	return &TrackEncoder{
		vc:         gs.voiceConnection,
		sources:    p.sources,
		ffmpegPath: p.ffmpegPath,
		log:        p.log,
		onStop: func(_ player.StopMessage) {
			gs.mu.Lock()
			vc := gs.voiceConnection
			gs.mu.Unlock()

			if vc != nil {
				_ = vc.Disconnect()
			}
			p.removeGuildState(guildID)
		},
	}
}

func (p *Manager) iniGuildState(ctx context.Context, guildID, textChannelID string, userID string) (*guildState, error) {
	voiceState, err := p.session.UserVoiceState(guildID, userID, discordgo.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	if voiceState.ChannelID == "" {
		return nil, ErrNotInVoiceChannel
	}

	gs := p.resolveGuildState(guildID, textChannelID)
	if gs.voiceConnection != nil {
		if gs.voiceConnection.ChannelID == voiceState.ChannelID {
			return gs, nil
		}
		if err := gs.voiceConnection.ChangeChannel(voiceState.ChannelID, false, true); err != nil {
			return nil, err
		}
		return gs, nil
	}

	vc, err := p.session.ChannelVoiceJoin(guildID, voiceState.ChannelID, false, true)
	if err != nil {
		return nil, fmt.Errorf("failed to join voice channel: %w", err)
	}
	gs.voiceConnection = vc

	return gs, nil
}

func (p *Manager) resolveTrack(ctx context.Context, userID, url, sourceName string) (*player.Track, error) {
	src, ok := p.sources.Find(sourceName)
	if !ok {
		return nil, fmt.Errorf("source %q not found", sourceName)
	}

	track, err := src.Resolve(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve track: %w", err)
	}

	return &player.Track{
		Title:       track.Title,
		URL:         track.URL,
		Duration:    track.Duration,
		RequestedBy: userID,
		Source:      sourceName,
	}, nil
}

func (p *Manager) resolveGuildState(guildID, textChannelID string) *guildState {
	p.mu.Lock()
	defer p.mu.Unlock()

	gs, ok := p.guildStates[guildID]
	if !ok {
		gs = &guildState{
			guildID:       guildID,
			textChannelID: textChannelID,
			player:        player.New(p.idleTimeout, p.log.Desugar()),
		}
		p.guildStates[guildID] = gs
	}
	return gs
}

func (p *Manager) removeGuildState(guildID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.guildStates, guildID)
}

func (p *Manager) withGuildState(guildID string, cb func(gs *guildState) error) error {
	p.mu.RLock()
	gs, ok := p.guildStates[guildID]
	p.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%w: %s", ErrNotInitialized, guildID)
	}
	return cb(gs)
}
