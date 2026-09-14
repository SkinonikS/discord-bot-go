package autorole

import (
	"context"
	"errors"
	"fmt"

	autorole "github.com/SkinonikS/discord-bot-go/internal/service/repository/auto_role"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	ErrAutoRoleAlreadyExists = errors.New("auto role already exists")
	ErrAutoRoleNotFound      = errors.New("auto role not found")
)

type Service interface {
	AddAutoRole(ctx context.Context, params AddAutoRole) (*autorole.AutoRole, error)
	RemoveAutoRole(ctx context.Context, params RemoveAutoRole) error
	ListAutoRoles(ctx context.Context, guildID snowflake.ID) ([]autorole.AutoRole, error)
	RoleDelete(ctx context.Context, params RoleDelete) error
	GuildMemberJoin(ctx context.Context, params GuildMemberJoin) error
}

type serviceImpl struct {
	autoRoleRepo autorole.Repo
	discordApi   disgorest.Rest
	log          *zap.SugaredLogger
}

type ServiceParams struct {
	fx.In

	Log          *zap.Logger
	DiscordApi   disgorest.Rest
	AutoRoleRepo autorole.Repo
}

func NewService(p ServiceParams) Service {
	return &serviceImpl{
		discordApi:   p.DiscordApi,
		autoRoleRepo: p.AutoRoleRepo,
		log:          p.Log.Sugar(),
	}
}

type AddAutoRole struct {
	GuildID snowflake.ID
	RoleID  snowflake.ID
}

func (s *serviceImpl) AddAutoRole(ctx context.Context, params AddAutoRole) (*autorole.AutoRole, error) {
	existing, err := s.autoRoleRepo.FindByGuildID(ctx, params.GuildID)
	if err != nil {
		return nil, fmt.Errorf("failed to find auto roles: %w", err)
	}
	for _, autoRole := range existing {
		if autoRole.RoleID == params.RoleID {
			return nil, ErrAutoRoleAlreadyExists
		}
	}

	autoRole := &autorole.AutoRole{
		GuildID: params.GuildID,
		RoleID:  params.RoleID,
	}
	if err := s.autoRoleRepo.Save(ctx, autoRole); err != nil {
		return nil, fmt.Errorf("failed to save auto role: %w", err)
	}

	return autoRole, nil
}

type RemoveAutoRole struct {
	GuildID snowflake.ID
	RoleID  snowflake.ID
}

func (s *serviceImpl) RemoveAutoRole(ctx context.Context, params RemoveAutoRole) error {
	count, err := s.autoRoleRepo.DeleteByGuildIDAndRoleID(ctx, params.GuildID, params.RoleID)
	if err != nil {
		return fmt.Errorf("failed to remove auto role: %w", err)
	}
	if count == 0 {
		return ErrAutoRoleNotFound
	}

	return nil
}

func (s *serviceImpl) ListAutoRoles(ctx context.Context, guildID snowflake.ID) ([]autorole.AutoRole, error) {
	return s.autoRoleRepo.FindByGuildID(ctx, guildID)
}

type RoleDelete struct {
	GuildID snowflake.ID
	RoleID  snowflake.ID
}

func (s *serviceImpl) RoleDelete(ctx context.Context, params RoleDelete) error {
	_, err := s.autoRoleRepo.DeleteByGuildIDAndRoleID(ctx, params.GuildID, params.RoleID)
	return err
}

type GuildMemberJoin struct {
	GuildID snowflake.ID
	UserID  snowflake.ID
}

func (s *serviceImpl) GuildMemberJoin(ctx context.Context, params GuildMemberJoin) error {
	autoRoles, err := s.autoRoleRepo.FindByGuildID(ctx, params.GuildID)
	if err != nil {
		return fmt.Errorf("failed to find auto roles: %w", err)
	}

	var errs error
	for _, autoRole := range autoRoles {
		if err := s.discordApi.AddMemberRole(params.GuildID, params.UserID, autoRole.RoleID, disgorest.WithCtx(ctx)); err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to add auto role %d: %w", autoRole.RoleID, err))
		}
	}

	return errs
}
