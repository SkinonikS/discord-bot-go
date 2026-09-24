package auto_role

import (
	"context"
	"fmt"
	"uuid"

	"github.com/SkinonikS/discord-bot-go/internal/service/repository/auto_role"
	"github.com/SkinonikS/discord-bot-go/internal/service/repository/postgres/internal/gen"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var _ autorole.Repo = (*Repo)(nil)

type Repo struct {
	queries *gen.Queries
}

type Params struct {
	fx.In

	Pool *pgxpool.Pool
}

func NewRepo(p Params) *Repo {
	return &Repo{
		queries: gen.New(p.Pool),
	}
}

func (r *Repo) Find(ctx context.Context, findParams autorole.FindParams) ([]autorole.AutoRole, error) {
	rawAutoRoles, err := r.queries.FindAutoRoles(ctx, gen.FindAutoRolesParams{
		GuildID: findParams.GuildID,
		RoleID:  findParams.RoleID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find auto roles: %w", err)
	}

	autoRoles := make([]autorole.AutoRole, len(rawAutoRoles))
	for i, rawAutoRole := range rawAutoRoles {
		autoRoles[i] = autorole.AutoRole{
			ID:      rawAutoRole.ID,
			GuildID: rawAutoRole.GuildID,
			RoleID:  rawAutoRole.RoleID,
		}
	}

	return autoRoles, nil
}

func (r *Repo) Save(ctx context.Context, autoRole *autorole.AutoRole) error {
	if autoRole.ID == uuid.Nil() {
		autoRole.ID = uuid.New()
	}

	return r.queries.SaveAutoRole(ctx, gen.SaveAutoRoleParams{
		ID:      autoRole.ID,
		GuildID: autoRole.GuildID,
		RoleID:  autoRole.RoleID,
	})
}

func (r *Repo) Delete(ctx context.Context, deleteParams autorole.DeleteParams) (int64, error) {
	return r.queries.DeleteAutoRoles(ctx, gen.DeleteAutoRolesParams{
		GuildID: deleteParams.GuildID,
		RoleID:  deleteParams.RoleID,
	})
}
