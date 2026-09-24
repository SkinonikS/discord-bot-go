package temp_voice_channel_state

import (
	"context"
	"errors"
	"uuid"

	"github.com/SkinonikS/discord-bot-go/internal/service/repository/postgres/internal/gen"
	"github.com/SkinonikS/discord-bot-go/internal/service/repository/temp_voice_channel_state"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var _ tempvoicechannelstate.Repo = (*Repo)(nil)

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

func (r *Repo) Transaction(ctx context.Context, fn func(tx tempvoicechannelstate.Repo) error) error {
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

func (r *Repo) Find(ctx context.Context, params tempvoicechannelstate.FindParams) ([]tempvoicechannelstate.TempVoiceChannelState, error) {
	rawStates, err := r.queries.FindTempVoiceChannelStatesByCriteria(ctx, gen.FindTempVoiceChannelStatesByCriteriaParams{
		GuildID:   params.GuildID,
		ChannelID: params.ChannelID,
	})
	if err != nil {
		return nil, err
	}

	states := make([]tempvoicechannelstate.TempVoiceChannelState, len(rawStates))
	for i, rawState := range rawStates {
		states[i] = tempvoicechannelstate.TempVoiceChannelState{
			ID:        rawState.ID,
			GuildID:   rawState.GuildID,
			ChannelID: rawState.ChannelID,
		}
	}

	return states, nil
}

func (r *Repo) Save(ctx context.Context, state *tempvoicechannelstate.TempVoiceChannelState) error {
	if state.ID == uuid.Nil() {
		state.ID = uuid.New()
	}

	return r.queries.SaveTempVoiceChannelState(ctx, gen.SaveTempVoiceChannelStateParams{
		ID:        state.ID,
		GuildID:   state.GuildID,
		ChannelID: state.ChannelID,
	})
}

func (r *Repo) DeleteManyByIDs(ctx context.Context, ids []uuid.UUID) (int64, error) {
	return r.queries.DeleteTempVoiceChannelStatesByIDs(ctx, ids)
}
