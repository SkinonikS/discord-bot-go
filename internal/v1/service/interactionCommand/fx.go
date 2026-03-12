package interactionCommand

import (
	"context"

	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	disgobot "github.com/disgoorg/disgo/bot"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "interactionCommand"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewRegistry),
		fx.Provide(
			AsCommand(NewPingCommand),
			AsCommand(NewInfoCommand),
			discord.AsEventListener(NewEventListener),
		),
		fx.Invoke(
			func(lc fx.Lifecycle, log *zap.Logger, discordConfig *discord.Config, discordClient *disgobot.Client, registry Registry) {
				lc.Append(fx.StartHook(
					func(ctx context.Context) error {
						go func() {
							definitions := lo.Map(registry.List(), func(cmd Command, _ int) disgodiscord.ApplicationCommandCreate {
								return cmd.Definition()
							})

							if _, err := discordClient.Rest.SetGlobalCommands(discordConfig.AppID, definitions, disgorest.WithCtx(ctx)); err != nil {
								log.Warn("failed to register global commands", zap.Error(err))
							}

							log.Info("global commands registered successfully", zap.Int("count", len(definitions)))
						}()

						return nil
					},
				))
			},
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}

func AsCommand(f any) any {
	return fx.Annotate(f, fx.As(new(Command)), fx.ResultTags(`group:"discord_commands"`))
}
