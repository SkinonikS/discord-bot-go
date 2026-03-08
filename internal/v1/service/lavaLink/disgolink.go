package lavaLink

import (
	"context"
	"log/slog"

	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/lavaqueue-plugin"
	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In
	EventListeners []disgolink.EventListener `group:"lavalink_event_listeners"`
	Log            *zap.Logger
	Lc             fx.Lifecycle
	Config         *Config
	Client         *disgobot.Client
}

func New(p Params) disgolink.Client {
	logger := slog.New(slogzap.Option{Logger: p.Log, Level: slog.LevelInfo}.NewZapHandler())
	lavaLink := disgolink.New(p.Client.ApplicationID,
		disgolink.WithLogger(logger),
		disgolink.WithPlugins(lavaqueue.NewWithLogger(logger)),
		disgolink.WithListeners(p.EventListeners...),
	)

	p.Lc.Append(fx.StartStopHook(
		func(ctx context.Context) error {
			go func() {
				for _, nodeConfig := range p.Config.Nodes {
					node, err := lavaLink.AddNode(ctx, disgolink.NodeConfig{
						Name:     nodeConfig.Name,
						Address:  nodeConfig.Address,
						Password: nodeConfig.Password,
						Secure:   nodeConfig.Secure,
					})
					if err != nil {
						p.Log.Warn("failed to add lavaLink node", zap.Error(err))
					}

					version, err := node.Version(ctx)
					if err != nil {
						p.Log.Warn("failed to get lavaLink node version", zap.Error(err), zap.String("node", nodeConfig.Name))
					}
					p.Log.Debug("connected to lavaLink node", zap.String("node", nodeConfig.Name), zap.String("version", version))
				}
			}()
			return nil
		},
		func(ctx context.Context) error {
			lavaLink.Close()
			p.Log.Info("lavaLink client closed")
			return nil
		},
	))

	return lavaLink
}
