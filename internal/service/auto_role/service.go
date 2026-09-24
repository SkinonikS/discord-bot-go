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
	AddAutoRole(ctx context.Context, params AddAutoRoleParams) (*autorole.AutoRole, error)
	ListAutoRoles(ctx context.Context, params ListAutoRolesParams) ([]autorole.AutoRole, error)
	RemoveAutoRole(ctx context.Context, params RemoveAutoRoleParams) error
	RoleDelete(ctx context.Context, params RoleDeleteParams) error
	GuildMemberJoin(ctx context.Context, params GuildMemberJoinParams) error
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

type AddAutoRoleParams struct {
	GuildID snowflake.ID
	RoleID  snowflake.ID
}

func (s *serviceImpl) AddAutoRole(ctx context.Context, params AddAutoRoleParams) (*autorole.AutoRole, error) {
	existing, err := s.autoRoleRepo.Find(ctx, autorole.FindParams{
		GuildID: params.GuildID,
		RoleID:  params.RoleID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find auto roles: %w", err)
	}
	if len(existing) > 0 {
		return nil, ErrAutoRoleAlreadyExists
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

type RemoveAutoRoleParams struct {
	GuildID snowflake.ID
	RoleID  snowflake.ID
}

func (s *serviceImpl) RemoveAutoRole(ctx context.Context, params RemoveAutoRoleParams) error {
	count, err := s.autoRoleRepo.Delete(ctx, autorole.DeleteParams{
		GuildID: params.GuildID,
		RoleID:  params.RoleID,
	})
	if err != nil {
		return fmt.Errorf("failed to remove auto role: %w", err)
	}
	if count == 0 {
		return ErrAutoRoleNotFound
	}

	return nil
}

type ListAutoRolesParams struct {
	GuildID snowflake.ID
}

func (s *serviceImpl) ListAutoRoles(ctx context.Context, params ListAutoRolesParams) ([]autorole.AutoRole, error) {
	return s.autoRoleRepo.Find(ctx, autorole.FindParams{
		GuildID: params.GuildID,
	})
}

type RoleDeleteParams struct {
	GuildID snowflake.ID
	RoleID  snowflake.ID
}

func (s *serviceImpl) RoleDelete(ctx context.Context, params RoleDeleteParams) error {
	_, err := s.autoRoleRepo.Delete(ctx, autorole.DeleteParams{
		GuildID: params.GuildID,
		RoleID:  params.RoleID,
	})
	return err
}

type GuildMemberJoinParams struct {
	GuildID snowflake.ID
	UserID  snowflake.ID
}

func (s *serviceImpl) GuildMemberJoin(ctx context.Context, params GuildMemberJoinParams) error {
	autoRoles, err := s.autoRoleRepo.Find(ctx, autorole.FindParams{
		GuildID: params.GuildID,
	})
	if err != nil {
		return fmt.Errorf("failed to find auto roles: %w", err)
	}

	var errs error
	for _, autoRole := range autoRoles {
		if err := s.discordApi.AddMemberRole(params.GuildID, params.UserID, autoRole.RoleID, disgorest.WithCtx(ctx)); err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to add auto role %d: %w", autoRole.RoleID, err))
		} else {
			s.log.Debugw("assigned auto role to user",
				zap.String("user_id", params.UserID.String()),
				zap.String("guild_id", params.GuildID.String()),
				zap.String("role_id", autoRole.RoleID.String()),
			)
		}
	}

	return errs
}
