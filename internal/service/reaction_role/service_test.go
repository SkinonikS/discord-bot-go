package reactionrole_test

import (
	"context"
	"errors"
	"testing"
	"uuid"

	discordmocks "github.com/SkinonikS/discord-bot-go/internal/infra/discord/mock"
	reactionrole "github.com/SkinonikS/discord-bot-go/internal/service/reaction_role"
	reactionrolerepo "github.com/SkinonikS/discord-bot-go/internal/service/repository/reaction_role"
	reactionrolemocks "github.com/SkinonikS/discord-bot-go/internal/service/repository/reaction_role/mock"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

const (
	testGuildID   = snowflake.ID(1)
	testChannelID = snowflake.ID(2)
	testMessageID = snowflake.ID(3)
	testRoleID    = snowflake.ID(4)
	testUserID    = snowflake.ID(5)
)

func newService(t *testing.T) (reactionrole.Service, *reactionrolemocks.MockRepo, *discordmocks.MockRest) {
	t.Helper()

	repo := reactionrolemocks.NewMockRepo(t)
	rest := discordmocks.NewMockRest(t)

	svc := reactionrole.NewService(reactionrole.ServiceParams{
		Log:              zap.NewNop(),
		DiscordApi:       rest,
		ReactionRoleRepo: repo,
	})

	return svc, repo, rest
}

// withTransaction makes the mocked Transaction call its callback with the same repo mock,
// mirroring how the real GORM-backed repo passes a tx-scoped Repo to fn.
func withTransaction(repo *reactionrolemocks.MockRepo) {
	repo.EXPECT().Transaction(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(tx reactionrolerepo.Repo) error) error {
		return fn(repo)
	})
}

func TestGuildMessageReactionRemoveEmoji(t *testing.T) {
	svc, repo, _ := newService(t)

	id := uuid.New()
	repo.EXPECT().Find(mock.Anything, reactionrolerepo.FindParams{
		GuildID:   testGuildID,
		ChannelID: testChannelID,
		MessageID: testMessageID,
		EmojiName: "pepe",
	}).Return([]reactionrolerepo.ReactionRole{{ID: id}}, nil)
	repo.EXPECT().DeleteManyByIDs(mock.Anything, []uuid.UUID{id}).Return(int64(1), nil)

	err := svc.GuildMessageReactionRemoveEmoji(context.Background(), reactionrole.GuildMessageReactionRemoveEmoji{
		GuildID:   testGuildID,
		ChannelID: testChannelID,
		MessageID: testMessageID,
		EmojiName: "pepe",
	})

	assert.NoError(t, err)
}

func TestGuildMessageReactionRemoveEmoji_FindError(t *testing.T) {
	svc, repo, _ := newService(t)

	repo.EXPECT().Find(mock.Anything, mock.Anything).Return(nil, errors.New("db down"))

	err := svc.GuildMessageReactionRemoveEmoji(context.Background(), reactionrole.GuildMessageReactionRemoveEmoji{GuildID: testGuildID})

	assert.ErrorContains(t, err, "db down")
}

func TestGuildMessageReactionRemoveAll(t *testing.T) {
	svc, repo, _ := newService(t)

	id := uuid.New()
	repo.EXPECT().Find(mock.Anything, reactionrolerepo.FindParams{
		GuildID:   testGuildID,
		ChannelID: testChannelID,
		MessageID: testMessageID,
	}).Return([]reactionrolerepo.ReactionRole{{ID: id}}, nil)
	repo.EXPECT().DeleteManyByIDs(mock.Anything, []uuid.UUID{id}).Return(int64(1), nil)

	err := svc.GuildMessageReactionRemoveAll(context.Background(), reactionrole.GuildMessageReactionRemoveAll{
		GuildID:   testGuildID,
		ChannelID: testChannelID,
		MessageID: testMessageID,
	})

	assert.NoError(t, err)
}

func TestGuildMessageDelete(t *testing.T) {
	svc, repo, _ := newService(t)

	repo.EXPECT().Find(mock.Anything, reactionrolerepo.FindParams{
		GuildID:   testGuildID,
		ChannelID: testChannelID,
		MessageID: testMessageID,
	}).Return(nil, nil)
	repo.EXPECT().DeleteManyByIDs(mock.Anything, []uuid.UUID{}).Return(int64(0), nil)

	err := svc.GuildMessageDelete(context.Background(), reactionrole.GuildMessageDelete{
		GuildID:   testGuildID,
		ChannelID: testChannelID,
		MessageID: testMessageID,
	})

	assert.NoError(t, err)
}

func TestRoleDelete(t *testing.T) {
	svc, repo, _ := newService(t)

	id := uuid.New()
	repo.EXPECT().Find(mock.Anything, reactionrolerepo.FindParams{
		GuildID: testGuildID,
		RoleID:  testRoleID,
	}).Return([]reactionrolerepo.ReactionRole{{ID: id}}, nil)
	repo.EXPECT().DeleteManyByIDs(mock.Anything, []uuid.UUID{id}).Return(int64(1), nil)

	err := svc.RoleDelete(context.Background(), reactionrole.RoleDelete{GuildID: testGuildID, RoleID: testRoleID})

	assert.NoError(t, err)
}

func TestGuildMessageReactionAdd(t *testing.T) {
	t.Run("adds the role when a reaction role matches", func(t *testing.T) {
		svc, repo, rest := newService(t)

		repo.EXPECT().Find(mock.Anything, reactionrolerepo.FindParams{
			GuildID:   testGuildID,
			MessageID: testMessageID,
			EmojiName: "pepe",
		}).Return([]reactionrolerepo.ReactionRole{{RoleID: testRoleID}}, nil)
		rest.EXPECT().AddMemberRole(testGuildID, testUserID, testRoleID, mock.Anything).Return(nil)

		err := svc.GuildMessageReactionAdd(context.Background(), reactionrole.GuildMessageReactionAdd{
			GuildID:   testGuildID,
			MessageID: testMessageID,
			UserID:    testUserID,
			EmojiName: "pepe",
		})

		assert.NoError(t, err)
	})

	t.Run("returns ErrNotReactionRole when nothing matches", func(t *testing.T) {
		svc, repo, _ := newService(t)

		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(nil, nil)

		err := svc.GuildMessageReactionAdd(context.Background(), reactionrole.GuildMessageReactionAdd{GuildID: testGuildID})

		assert.ErrorIs(t, err, reactionrole.ErrNotReactionRole)
	})

	t.Run("wraps the discord api error", func(t *testing.T) {
		svc, repo, rest := newService(t)

		repo.EXPECT().Find(mock.Anything, mock.Anything).Return([]reactionrolerepo.ReactionRole{{RoleID: testRoleID}}, nil)
		rest.EXPECT().AddMemberRole(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("forbidden"))

		err := svc.GuildMessageReactionAdd(context.Background(), reactionrole.GuildMessageReactionAdd{GuildID: testGuildID})

		assert.ErrorContains(t, err, "forbidden")
	})
}

func TestGuildMessageReactionRemove(t *testing.T) {
	t.Run("removes the role when a reaction role matches", func(t *testing.T) {
		svc, repo, rest := newService(t)

		repo.EXPECT().Find(mock.Anything, reactionrolerepo.FindParams{
			GuildID:   testGuildID,
			MessageID: testMessageID,
			EmojiName: "pepe",
		}).Return([]reactionrolerepo.ReactionRole{{RoleID: testRoleID}}, nil)
		rest.EXPECT().RemoveMemberRole(testGuildID, testUserID, testRoleID, mock.Anything).Return(nil)

		err := svc.GuildMessageReactionRemove(context.Background(), reactionrole.GuildMessageReactionRemove{
			GuildID:   testGuildID,
			MessageID: testMessageID,
			UserID:    testUserID,
			EmojiName: "pepe",
		})

		assert.NoError(t, err)
	})

	t.Run("returns ErrNotReactionRole when nothing matches", func(t *testing.T) {
		svc, repo, _ := newService(t)

		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(nil, nil)

		err := svc.GuildMessageReactionRemove(context.Background(), reactionrole.GuildMessageReactionRemove{GuildID: testGuildID})

		assert.ErrorIs(t, err, reactionrole.ErrNotReactionRole)
	})
}

func TestCreateReactionRole(t *testing.T) {
	t.Run("saves the role and reacts to the message with a custom emoji", func(t *testing.T) {
		svc, repo, rest := newService(t)

		withTransaction(repo)
		repo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(r *reactionrolerepo.ReactionRole) bool {
			return r.EmojiName == "pepe:123456789012345678"
		})).Return(nil)
		rest.EXPECT().AddReaction(testChannelID, testMessageID, "pepe:123456789012345678", mock.Anything).Return(nil)

		result, err := svc.CreateReactionRole(context.Background(), reactionrole.CreateReactionRole{
			GuildID:   testGuildID,
			ChannelID: testChannelID,
			MessageID: testMessageID,
			RoleID:    testRoleID,
			EmojiName: "<:pepe:123456789012345678>",
		})

		assert.NoError(t, err)
		assert.Equal(t, "pepe:123456789012345678", result.EmojiName)
	})

	t.Run("trims a plain unicode emoji instead of reformatting it", func(t *testing.T) {
		svc, repo, rest := newService(t)

		withTransaction(repo)
		repo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(r *reactionrolerepo.ReactionRole) bool {
			return r.EmojiName == "😀"
		})).Return(nil)
		rest.EXPECT().AddReaction(mock.Anything, mock.Anything, "😀", mock.Anything).Return(nil)

		_, err := svc.CreateReactionRole(context.Background(), reactionrole.CreateReactionRole{EmojiName: "  😀  "})

		assert.NoError(t, err)
	})

	t.Run("does not react when saving fails", func(t *testing.T) {
		svc, repo, _ := newService(t)

		withTransaction(repo)
		repo.EXPECT().Save(mock.Anything, mock.Anything).Return(errors.New("write failed"))

		_, err := svc.CreateReactionRole(context.Background(), reactionrole.CreateReactionRole{EmojiName: "😀"})

		assert.ErrorContains(t, err, "write failed")
	})

	t.Run("propagates the discord api error", func(t *testing.T) {
		svc, repo, rest := newService(t)

		withTransaction(repo)
		repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)
		rest.EXPECT().AddReaction(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("forbidden"))

		_, err := svc.CreateReactionRole(context.Background(), reactionrole.CreateReactionRole{EmojiName: "😀"})

		assert.ErrorContains(t, err, "forbidden")
	})
}

func TestDeleteReactionRole(t *testing.T) {
	t.Run("removes the reaction and the stored role", func(t *testing.T) {
		svc, repo, rest := newService(t)

		id := uuid.New()
		repo.EXPECT().Find(mock.Anything, reactionrolerepo.FindParams{
			GuildID:   testGuildID,
			MessageID: testMessageID,
			EmojiName: "pepe",
		}).Return([]reactionrolerepo.ReactionRole{{ID: id, ChannelID: testChannelID}}, nil)
		withTransaction(repo)
		repo.EXPECT().DeleteManyByIDs(mock.Anything, []uuid.UUID{id}).Return(int64(1), nil)
		rest.EXPECT().RemoveOwnReaction(testChannelID, testMessageID, "pepe", mock.Anything).Return(nil)

		err := svc.DeleteReactionRole(context.Background(), reactionrole.DeleteReactionRole{
			GuildID:   testGuildID,
			MessageID: testMessageID,
			EmojiName: "pepe",
		})

		assert.NoError(t, err)
	})

	t.Run("returns ErrReactionRoleNotFound when nothing matches", func(t *testing.T) {
		svc, repo, _ := newService(t)

		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(nil, nil)

		err := svc.DeleteReactionRole(context.Background(), reactionrole.DeleteReactionRole{GuildID: testGuildID})

		assert.ErrorIs(t, err, reactionrole.ErrReactionRoleNotFound)
	})

	t.Run("wraps the lookup error", func(t *testing.T) {
		svc, repo, _ := newService(t)

		repo.EXPECT().Find(mock.Anything, mock.Anything).Return(nil, errors.New("db down"))

		err := svc.DeleteReactionRole(context.Background(), reactionrole.DeleteReactionRole{GuildID: testGuildID})

		assert.ErrorContains(t, err, "db down")
	})
}
