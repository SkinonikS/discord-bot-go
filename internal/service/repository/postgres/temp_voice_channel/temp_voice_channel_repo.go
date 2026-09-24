package postgres

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/SkinonikS/discord-bot-go/internal/service/repository/postgres/internal/gen"
	"github.com/SkinonikS/discord-bot-go/internal/service/repository/temp_voice_channel"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

var _ tempvoicechannel.Repo = (*Repo)(nil)

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

func (c *Repo) Transaction(ctx context.Context, fn func(tx tempvoicechannel.Repo) error) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return err
	}

	err = fn(&Repo{
		queries: c.queries.WithTx(tx),
		pool:    c.pool,
	})
	if err != nil {
		return errors.Join(err, tx.Rollback(ctx))
	}

	return tx.Commit(ctx)
}

func (c *Repo) Find(ctx context.Context, params tempvoicechannel.FindParams) ([]tempvoicechannel.TempVoiceChannel, error) {
	rawChannels, err := c.queries.FindTempVoiceChannels(ctx, gen.FindTempVoiceChannelsParams{
		GuildID:       params.GuildID,
		RootChannelID: params.RootChannelID,
		ParentID:      params.ParentID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find temp voice channels: %w", err)
	}

	channels := make([]tempvoicechannel.TempVoiceChannel, len(rawChannels))
	for i, rawChannel := range rawChannels {
		channels[i] = tempvoicechannel.TempVoiceChannel{
			ID:            rawChannel.ID,
			GuildID:       rawChannel.GuildID,
			RootChannelID: rawChannel.RootChannelID,
			ParentID:      rawChannel.ParentID,
		}
	}

	return channels, nil
}

func (c *Repo) Save(ctx context.Context, setupChannel *tempvoicechannel.TempVoiceChannel) error {
	if setupChannel.ID == uuid.Nil() {
		setupChannel.ID = uuid.New()
	}

	return c.queries.SaveTempVoiceChannel(ctx, gen.SaveTempVoiceChannelParams{
		ID:            setupChannel.ID,
		GuildID:       setupChannel.GuildID,
		RootChannelID: setupChannel.RootChannelID,
		ParentID:      setupChannel.ParentID,
	})
}

func (c *Repo) DeleteManyByIDs(ctx context.Context, ids []uuid.UUID) (int64, error) {
	return c.queries.DeleteTempVoiceChannelsByIDs(ctx, ids)
}
