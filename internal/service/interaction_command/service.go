package interactioncommand

import (
	"context"
	"errors"
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/infra/discord"
	disabledguildcommands "github.com/SkinonikS/discord-bot-go/internal/service/repository/disabled_guild_commands"
	disgobot "github.com/disgoorg/disgo/bot"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const guildPageSize = 200

type Service interface {
	IsGuildCommandDisabled(ctx context.Context, params IsGuildCommandDisabledParams) (bool, error)
	ListDisabledGuildCommands(ctx context.Context, guildID snowflake.ID) ([]string, error)
	SetGuildCommandState(ctx context.Context, params SetGuildCommandStateParams) error
	SyncGlobalCommands(ctx context.Context) error
	SyncGuildCommands(ctx context.Context, guildID snowflake.ID) error
	SyncAllGuildCommands(ctx context.Context) error
}

type serviceImpl struct {
	log                       *zap.SugaredLogger
	config                    *discord.Config
	client                    *disgobot.Client
	disabledGuildCommandsRepo disabledguildcommands.Repo
	registry                  Registry
}

type ServiceParams struct {
	fx.In

	Log                       *zap.Logger
	Config                    *discord.Config
	Client                    *disgobot.Client
	DisabledGuildCommandsRepo disabledguildcommands.Repo
	Registry                  Registry
}

func NewService(p ServiceParams) Service {
	return &serviceImpl{
		log:                       p.Log.Sugar(),
		config:                    p.Config,
		client:                    p.Client,
		registry:                  p.Registry,
		disabledGuildCommandsRepo: p.DisabledGuildCommandsRepo,
	}
}

type IsGuildCommandDisabledParams struct {
	GuildID     snowflake.ID
	CommandName string
}

func (s *serviceImpl) IsGuildCommandDisabled(ctx context.Context, params IsGuildCommandDisabledParams) (bool, error) {
	return s.disabledGuildCommandsRepo.IsDisabled(ctx, params.GuildID, params.CommandName)
}

func (s *serviceImpl) ListDisabledGuildCommands(ctx context.Context, guildID snowflake.ID) ([]string, error) {
	return s.disabledGuildCommandsRepo.ListDisabled(ctx, guildID)
}

type SetGuildCommandStateParams struct {
	GuildID snowflake.ID
	Command string
	Disable bool
}

func (s *serviceImpl) SetGuildCommandState(ctx context.Context, params SetGuildCommandStateParams) error {
	if params.Disable {
		return s.disabledGuildCommandsRepo.Disable(ctx, params.GuildID, params.Command)
	}

	return s.disabledGuildCommandsRepo.Enable(ctx, params.GuildID, params.Command)
}

func (s *serviceImpl) SyncGlobalCommands(ctx context.Context) error {
	definitions := lo.Map(s.registry.ListByScope(CommandScopeGlobal), func(cmd Command, _ int) disgodiscord.ApplicationCommandCreate {
		return cmd.Definition()
	})

	if _, err := s.client.Rest.SetGlobalCommands(s.config.AppID, definitions, disgorest.WithCtx(ctx)); err != nil {
		return fmt.Errorf("failed to register global commands: %w", err)
	}

	s.log.Infow("global commands registered successfully", zap.Int("count", len(definitions)))
	return nil
}

func (s *serviceImpl) SyncGuildCommands(ctx context.Context, guildID snowflake.ID) error {
	disabled, err := s.disabledGuildCommandsRepo.ListDisabled(ctx, guildID)
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

func (s *serviceImpl) SyncAllGuildCommands(ctx context.Context) error {
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
