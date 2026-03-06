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
	disgobot "github.com/disgoorg/disgo/bot"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgovoice "github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	ErrNotInitialized    = errors.New("guild player not initialized")
	ErrNotInVoiceChannel = errors.New("you are not in a voice channel")
	ErrNotPlaying        = errors.New("not currently playing")
	ErrNotPaused         = errors.New("not currently paused")
	ErrSourceNotPlaylist = errors.New("source does not support playlists")
)

type guildState struct {
	mu              sync.Mutex
	guildID         snowflake.ID
	textChannelID   snowflake.ID
	player          *player.Player
	voiceConnection disgovoice.Conn
}

type PlayParams struct {
	GuildID       snowflake.ID
	UserID        snowflake.ID
	TextChannelID snowflake.ID
	URL           string
	SourceName    string
	PlayNow       bool
}

type PlaylistParams struct {
	GuildID       snowflake.ID
	UserID        snowflake.ID
	TextChannelID snowflake.ID
	URL           string
	SourceName    string
}

type Manager struct {
	mu          sync.RWMutex
	idleTimeout time.Duration
	client      *disgobot.Client
	log         *zap.SugaredLogger
	sources     *musicPlayerSource.Registry
	ffmpegPath  string
	guildStates map[snowflake.ID]*guildState
}

type ManagerParams struct {
	fx.In
	Config  *Config
	Path    *foundation.Path
	Client  *disgobot.Client
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
		client:      p.Client,
		ffmpegPath:  ffmpegPath,
		sources:     p.Sources,
		log:         p.Log.Sugar(),
		guildStates: make(map[snowflake.ID]*guildState),
	}, nil
}

func (p *Manager) Play(ctx context.Context, playParams PlayParams) (*player.Track, error) {
	gs, err := p.iniGuildState(ctx, playParams.GuildID, playParams.TextChannelID, playParams.UserID)
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

func (p *Manager) AddPlaylist(ctx context.Context, params PlaylistParams) ([]*player.Track, error) {
	gs, err := p.iniGuildState(ctx, params.GuildID, params.TextChannelID, params.UserID)
	if err != nil {
		return nil, err
	}

	src, ok := p.sources.Find(params.SourceName)
	if !ok {
		return nil, fmt.Errorf("source %q not found", params.SourceName)
	}

	ps, ok := src.(musicPlayerSource.PlaylistSource)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrSourceNotPlaylist, params.SourceName)
	}

	sourceTracks, err := ps.ResolvePlaylist(ctx, params.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve playlist: %w", err)
	}
	if len(sourceTracks) == 0 {
		return nil, errors.New("playlist is empty")
	}

	tracks := make([]*player.Track, 0, len(sourceTracks))
	for _, st := range sourceTracks {
		tracks = append(tracks, &player.Track{
			Title:       st.Title,
			URL:         st.URL,
			Duration:    st.Duration,
			RequestedBy: params.UserID.String(),
			Source:      params.SourceName,
		})
	}

	for _, t := range tracks {
		gs.player.PushQueue(t)
	}

	if gs.player.Status() == player.StatusIdle {
		go gs.player.Play(p.newTrackEncoder(gs, params.GuildID))
	}

	return tracks, nil
}

func (p *Manager) Skip(guildID snowflake.ID) error {
	return p.withGuildState(guildID, func(gs *guildState) error {
		gs.player.Skip()
		return nil
	})
}

func (p *Manager) Stop(guildID snowflake.ID) error {
	return p.withGuildState(guildID, func(gs *guildState) error {
		gs.player.Stop()
		return nil
	})
}

func (p *Manager) Remove(guildID snowflake.ID, index int) (*player.Track, error) {
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

func (p *Manager) Pause(guildID snowflake.ID) error {
	return p.withGuildState(guildID, func(gs *guildState) error {
		if !gs.player.Pause() {
			return ErrNotPlaying
		}
		return nil
	})
}

func (p *Manager) Resume(guildID snowflake.ID) error {
	return p.withGuildState(guildID, func(gs *guildState) error {
		if !gs.player.Resume() {
			return ErrNotPaused
		}
		return nil
	})
}

func (p *Manager) EraseQueue(guildID snowflake.ID) error {
	return p.withGuildState(guildID, func(gs *guildState) error {
		gs.player.EraseQueue()
		return nil
	})
}

func (p *Manager) Queue(guildID snowflake.ID) (*player.Queue, error) {
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

func (p *Manager) Disconnect(guildID snowflake.ID) error {
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
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		vc.Close(ctx)
	}

	p.removeGuildState(guildID)
	return nil
}

func (p *Manager) HasGuildState(guildID snowflake.ID) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.guildStates[guildID]
	return ok
}

func (p *Manager) newTrackEncoder(gs *guildState, guildID snowflake.ID) *TrackEncoder {
	return &TrackEncoder{
		vc:         gs.voiceConnection,
		sources:    p.sources,
		ffmpegPath: p.ffmpegPath,
		log:        p.log,
		onStop: func(msg player.StopMessage) {
			gs.mu.Lock()
			vc := gs.voiceConnection
			textChannelID := gs.textChannelID
			gs.mu.Unlock()

			if vc != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				vc.Close(ctx)
			}
			p.removeGuildState(guildID)

			var content string
			switch v := msg.Message.(type) {
			case error:
				content = v.Error()
			case string:
				content = v
			default:
				content = fmt.Sprintf("%v", msg.Message)
			}

			_, _ = p.client.Rest.CreateMessage(textChannelID, disgodiscord.MessageCreate{Content: content})
		},
	}
}

func (p *Manager) iniGuildState(ctx context.Context, guildID, textChannelID snowflake.ID, userID snowflake.ID) (*guildState, error) {
	vs, ok := p.client.Caches.VoiceState(guildID, userID)
	if !ok || vs.ChannelID == nil {
		return nil, ErrNotInVoiceChannel
	}
	channelID := *vs.ChannelID

	gs := p.resolveGuildState(guildID, textChannelID)

	gs.mu.Lock()
	vc := gs.voiceConnection
	gs.mu.Unlock()

	if vc != nil {
		currentChannelID := vc.ChannelID()
		if currentChannelID != nil && *currentChannelID == channelID {
			return gs, nil
		}
		if err := p.client.UpdateVoiceState(ctx, guildID, &channelID, false, true); err != nil {
			return nil, err
		}
		return gs, nil
	}

	conn := p.client.VoiceManager.CreateConn(guildID)
	if err := conn.Open(ctx, channelID, false, true); err != nil {
		p.client.VoiceManager.RemoveConn(guildID)
		return nil, fmt.Errorf("failed to join voice channel: %w", err)
	}

	gs.mu.Lock()
	gs.voiceConnection = conn
	gs.mu.Unlock()

	return gs, nil
}

func (p *Manager) resolveTrack(ctx context.Context, userID snowflake.ID, url, sourceName string) (*player.Track, error) {
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
		RequestedBy: userID.String(),
		Source:      sourceName,
	}, nil
}

func (p *Manager) resolveGuildState(guildID, textChannelID snowflake.ID) *guildState {
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

func (p *Manager) removeGuildState(guildID snowflake.ID) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.guildStates, guildID)
}

func (p *Manager) withGuildState(guildID snowflake.ID, cb func(gs *guildState) error) error {
	p.mu.Lock()
	gs, ok := p.guildStates[guildID]
	p.mu.Unlock()
	if !ok {
		return fmt.Errorf("%w: %s", ErrNotInitialized, guildID)
	}
	return cb(gs)
}
