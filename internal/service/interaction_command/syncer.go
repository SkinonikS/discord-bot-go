package interactioncommand

import (
	"context"
	"errors"
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/infra/discord"
	disgobot "github.com/disgoorg/disgo/bot"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const guildPageSize = 200

// Syncer pushes command definitions to Discord over REST.
type Syncer interface {
	// SyncGlobalCommands registers every CommandScopeGlobal command account-wide.
	SyncGlobalCommands(ctx context.Context) error
	// SyncGuildCommands registers CommandScopeGuild commands for a single guild, filtered
	// by that guild's disabled commands.
	SyncGuildCommands(ctx context.Context, guildID snowflake.ID) error
	// SyncAllGuildCommands runs SyncGuildCommands for every guild the bot is currently in.
	SyncAllGuildCommands(ctx context.Context) error
}

type syncerImpl struct {
	log      *zap.SugaredLogger
	config   *discord.Config
	client   *disgobot.Client
	registry Registry
	settings Settings
}

type SyncerParams struct {
	fx.In

	Log      *zap.Logger
	Config   *discord.Config
	Client   *disgobot.Client
	Registry Registry
	Settings Settings
}

func NewSyncer(p SyncerParams) Syncer {
	return &syncerImpl{
		log:      p.Log.Sugar(),
		config:   p.Config,
		client:   p.Client,
		registry: p.Registry,
		settings: p.Settings,
	}
}

func (s *syncerImpl) SyncGlobalCommands(ctx context.Context) error {
	definitions := lo.Map(s.registry.ListByScope(CommandScopeGlobal), func(cmd Command, _ int) disgodiscord.ApplicationCommandCreate {
		return cmd.Definition()
	})

	if _, err := s.client.Rest.SetGlobalCommands(s.config.AppID, definitions, disgorest.WithCtx(ctx)); err != nil {
		return fmt.Errorf("failed to register global commands: %w", err)
	}

	s.log.Infow("global commands registered successfully", zap.Int("count", len(definitions)))
	return nil
}

func (s *syncerImpl) SyncGuildCommands(ctx context.Context, guildID snowflake.ID) error {
	disabled, err := s.settings.ListDisabled(ctx, guildID)
	if err != nil {
		return fmt.Errorf("failed to list disabled commands for guild %s: %w", guildID, err)
	}
	disabledSet := lo.SliceToMap(disabled, func(name string) (string, struct{}) {
		return name, struct{}{}
	})

	commands := lo.Filter(s.registry.ListByScope(CommandScopeGuild), func(cmd Command, _ int) bool {
		_, ok := disabledSet[cmd.Name()]
		return !ok
	})

	definitions := lo.Map(commands, func(cmd Command, _ int) disgodiscord.ApplicationCommandCreate {
		return cmd.Definition()
	})

	if _, err := s.client.Rest.SetGuildCommands(s.config.AppID, guildID, definitions, disgorest.WithCtx(ctx)); err != nil {
		return fmt.Errorf("failed to register commands for guild %s: %w", guildID, err)
	}

	s.log.Infow("guild commands registered successfully", zap.String("guild_id", guildID.String()), zap.Int("count", len(definitions)))
	return nil
}

func (s *syncerImpl) SyncAllGuildCommands(ctx context.Context) error {
	var guildIDs []snowflake.ID

	page := s.client.Rest.GetCurrentUserGuildsPage("", 0, guildPageSize, false, disgorest.WithCtx(ctx))
	for page.Next() {
		for _, guild := range page.Items {
			guildIDs = append(guildIDs, guild.ID)
		}
	}
	if page.Err != nil && !errors.Is(page.Err, disgorest.ErrNoMorePages) {
		return fmt.Errorf("failed to list current guilds: %w", page.Err)
	}

	for _, guildID := range guildIDs {
		if err := s.SyncGuildCommands(ctx, guildID); err != nil {
			return err
		}
	}

	s.log.Infow("synced commands for all guilds", zap.Int("count", len(guildIDs)))
	return nil
}
