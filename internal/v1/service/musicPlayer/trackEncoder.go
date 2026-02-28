package musicPlayer

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/v1/service/musicPlayer/player"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/musicPlayerSource"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type TrackEncoder struct {
	vc         *discordgo.VoiceConnection
	sources    *musicPlayerSource.Registry
	ffmpegPath string
	log        *zap.SugaredLogger
	onStop     func(player.StopMessage)
}

func (e *TrackEncoder) Play(track *player.Track, event <-chan player.EventType) (player.EventType, error) {
	src, ok := e.sources.Find(track.Source)
	if !ok {
		return player.EventSkip, fmt.Errorf("source %q not found", track.Source)
	}

	stream, err := src.Stream(context.Background(), track.URL)
	if err != nil {
		return player.EventSkip, fmt.Errorf("failed to open stream: %w", err)
	}
	defer func() {
		if err := stream.Close(); err != nil {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				e.log.Errorw("failed to close stream", zap.Error(err))
			}
		}
	}()

	enc, err := player.NewFfmpegEncoder(stream, e.ffmpegPath)
	if err != nil {
		return player.EventSkip, fmt.Errorf("failed to create encoder: %w", err)
	}
	if err := enc.Start(); err != nil {
		return player.EventSkip, fmt.Errorf("failed to start ffmpeg: %w", err)
	}
	defer func() {
		if err := enc.Stop(); err != nil {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				e.log.Errorw("failed to stop ffmpeg", zap.Error(err))
			}
		}
	}()

	if err := e.vc.Speaking(true); err != nil {
		return player.EventSkip, fmt.Errorf("failed to set speaking: %w", err)
	}

	defer func() {
		if e.vc.Ready {
			if err := e.vc.Speaking(false); err != nil {
				e.log.Warnw("failed to set speaking: %w", zap.Error(err))
			}
		}
	}()

	frameCh := make(chan []byte, 2)
	doneCh := make(chan struct{})
	defer close(doneCh)

	pauseSig := make(chan struct{}, 1)
	resumeSig := make(chan struct{}, 1)

	encErrCh := make(chan error, 1)
	go func() {
		defer close(frameCh)
		for {
			frame, err := enc.Encode()
			if err != nil {
				encErrCh <- err
				return
			}
			if frame == nil {
				return
			}
			select {
			case frameCh <- frame:
			case <-doneCh:
				return
			}
		}
	}()

	senderDone := make(chan struct{})
	go func() {
		defer close(senderDone)
		paused := false
		for {
			if paused {
				select {
				case <-resumeSig:
					paused = false
				case <-doneCh:
					return
				}
				time.Sleep(100 * time.Millisecond)
				continue
			}
			select {
			case <-pauseSig:
				paused = true
			case frame, ok := <-frameCh:
				if !ok {
					return
				}
				if !e.vc.Ready {
					return
				}
				select {
				case e.vc.OpusSend <- frame:
				case <-doneCh:
					return
				}
			case <-doneCh:
				return
			}
		}
	}()

	for {
		select {
		case evt := <-event:
			switch evt {
			case player.EventStop:
				return player.EventStop, nil
			case player.EventSkip:
				return player.EventSkip, nil
			case player.EventPause:
				select {
				case pauseSig <- struct{}{}:
				default:
				}
			case player.EventResume:
				select {
				case resumeSig <- struct{}{}:
				default:
				}
			}
		case <-senderDone:
			select {
			case err := <-encErrCh:
				return player.EventStop, err
			default:
				return player.EventSkip, nil
			}
		}
	}
}

func (e *TrackEncoder) OnStop(msg player.StopMessage) {
	if e.onStop != nil {
		e.onStop(msg)
	}
}

func (e *TrackEncoder) OnError(err error) {
	e.log.Errorw("track encoder error", zap.Error(err))
}
