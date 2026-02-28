package youtubeMusicPlayerSource

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/musicPlayerSource"
	"go.uber.org/fx"
)

type Source struct {
	ytdlpPath string
	timeout   time.Duration
}

type Params struct {
	fx.In
	Config *Config
	Path   *foundation.Path
}

func New(p Params) (*Source, error) {
	ytdlpPath, err := p.Path.ResolveFrom(p.Config.YtdlpPath)
	if err != nil {
		return nil, err
	}

	return &Source{
		ytdlpPath: ytdlpPath,
		timeout:   p.Config.Timeout,
	}, nil
}

func (p *Source) Resolve(ctx context.Context, url string) (*musicPlayerSource.Track, error) {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	out, err := p.run(ctx,
		"--no-playlist",
		"--print", "%(title)s",
		"--print", "%(duration)s",
		url,
	)
	if err != nil {
		return nil, fmt.Errorf("yt-dlp: %w", err)
	}

	lines := strings.SplitN(strings.TrimSpace(out), "\n", 2)
	if len(lines) < 2 {
		return nil, fmt.Errorf("unexpected yt-dlp output: %q", out)
	}

	title := strings.TrimSpace(lines[0])
	durationSecs, _ := strconv.ParseFloat(strings.TrimSpace(lines[1]), 64)

	return &musicPlayerSource.Track{
		URL:      url,
		Title:    title,
		Duration: time.Duration(durationSecs) * time.Second,
	}, nil
}

func (p *Source) Stream(ctx context.Context, url string) (io.ReadCloser, error) {
	cmd := exec.CommandContext(ctx, p.ytdlpPath,
		"--no-playlist",
		"-f", "bestaudio/best",
		"-o", "-",
		url,
	)

	rc, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("yt-dlp start: %w", err)
	}

	return &Stream{rc: rc, cmd: cmd}, nil
}

func (p *Source) Name() string {
	return "YouTube"
}

func (p *Source) run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, p.ytdlpPath, args...)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("exit code %d: %s", exitErr.ExitCode(), strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}
