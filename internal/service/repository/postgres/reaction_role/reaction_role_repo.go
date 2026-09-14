package reaction_role

import (
	"context"
	"errors"
	"uuid"

	"github.com/SkinonikS/discord-bot-go/internal/service/repository/postgres/internal/gen"
	"github.com/SkinonikS/discord-bot-go/internal/service/repository/reaction_role"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var _ reactionrole.Repo = (*Repo)(nil)

type Repo struct {
	queries *gen.Queries
	pool    *pgxpool.Pool
}

type Params struct {
	fx.In

	Pool *pgxpool.Pool
}

func NewRepo(p Params) *Repo {
	return &Repo{
		queries: gen.New(p.Pool),
		pool:    p.Pool,
	}
}

func (r *Repo) Transaction(ctx context.Context, fn func(tx reactionrole.Repo) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	err = fn(&Repo{
		queries: r.queries.WithTx(tx),
		pool:    r.pool,
	})
	if err != nil {
		return errors.Join(err, tx.Rollback(ctx))
	}

	return tx.Commit(ctx)
}

func (r *Repo) FindByCriteria(ctx context.Context, criteria reactionrole.SearchCriteria) ([]reactionrole.ReactionRole, error) {
	rawReactionRoles, err := r.queries.FindReactionRolesByCriteria(ctx, gen.FindReactionRolesByCriteriaParams{
		GuildID:   criteria.GuildID,
		ChannelID: criteria.ChannelID,
		MessageID: criteria.MessageID,
		EmojiName: pgtype.Text{String: criteria.EmojiName, Valid: true},
		RoleID:    criteria.RoleID,
	})
	if err != nil {
		return nil, err
	}

	reactionRoles := make([]reactionrole.ReactionRole, len(rawReactionRoles))
	for i, rawReactionRole := range rawReactionRoles {
		reactionRoles[i] = reactionrole.ReactionRole{
			ID:        rawReactionRole.ID,
			GuildID:   rawReactionRole.GuildID,
			ChannelID: rawReactionRole.ChannelID,
			MessageID: rawReactionRole.MessageID,
			EmojiName: rawReactionRole.EmojiName,
			RoleID:    rawReactionRole.RoleID,
		}
	}

	return reactionRoles, nil
}

func (r *Repo) Save(ctx context.Context, reactionRole *reactionrole.ReactionRole) error {
	if reactionRole.ID == uuid.Nil() {
		reactionRole.ID = uuid.New()
	}

	id, err := r.queries.SaveReactionRole(ctx, gen.SaveReactionRoleParams{
		ID:        reactionRole.ID,
		GuildID:   reactionRole.GuildID,
		ChannelID: reactionRole.ChannelID,
		MessageID: reactionRole.MessageID,
		EmojiName: reactionRole.EmojiName,
		RoleID:    reactionRole.RoleID,
	})
	if err != nil {
		return err
	}

	reactionRole.ID = id
	return nil
}

func (r *Repo) DeleteManyByIDs(ctx context.Context, ids []uuid.UUID) (int64, error) {
	return r.queries.DeleteReactionRolesByIDs(ctx, ids)
}
