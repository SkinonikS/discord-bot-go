package lavaLink

import (
	"context"
	"log/slog"
	"net"
	"strconv"
	"time"

	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/lavaqueue-plugin"
	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In
	EventListeners []disgolink.EventListener `group:"lavaLink_event_listeners"`
	Log            *zap.Logger
	Lc             fx.Lifecycle
	Config         *Config
	Client         *disgobot.Client
}

func New(p Params) disgolink.Client {
	logger := slog.New(slogzap.Option{Logger: p.Log, Level: slog.LevelDebug}.NewZapHandler())
	lavaLink := disgolink.New(p.Client.ApplicationID,
		disgolink.WithLogger(logger),
		disgolink.WithPlugins(lavaqueue.NewWithLogger(logger)),
		disgolink.WithListeners(p.EventListeners...),
	)

	p.Lc.Append(fx.StartStopHook(
		func(context.Context) error {
			p.Log.Info("connecting to lavaLink nodes")
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
				defer cancel()

				for _, nodeConfig := range p.Config.Nodes {
					address := net.JoinHostPort(nodeConfig.Host, strconv.Itoa(nodeConfig.Port))
					node, err := lavaLink.AddNode(ctx, disgolink.NodeConfig{
						Name:     nodeConfig.Name,
						Address:  address,
						Password: nodeConfig.Password,
						Secure:   nodeConfig.Secure,
					})
					if err != nil {
						p.Log.Warn("failed to add lavaLink node", zap.Error(err))
						return
					}

					version, err := node.Version(ctx)
					if err != nil {
						p.Log.Warn("failed to get lavaLink node version", zap.Error(err), zap.String("node", nodeConfig.Name))
						return
					}
					p.Log.Info("connected to lavaLink node", zap.String("node", nodeConfig.Name), zap.String("version", version))
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
