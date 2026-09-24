package autorole_test

import (
	"errors"
	"testing"

	discordmocks "github.com/SkinonikS/discord-bot-go/internal/infra/discord/mock"
	autorole "github.com/SkinonikS/discord-bot-go/internal/service/auto_role"
	autorolerepo "github.com/SkinonikS/discord-bot-go/internal/service/repository/auto_role"
	autorolemocks "github.com/SkinonikS/discord-bot-go/internal/service/repository/auto_role/mock"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

const (
	testGuildID = snowflake.ID(1)
	testRoleID  = snowflake.ID(2)
	testUserID  = snowflake.ID(3)
)

func newService(t *testing.T) (autorole.Service, *autorolemocks.MockRepo, *discordmocks.MockRest) {
	t.Helper()

	repo := autorolemocks.NewMockRepo(t)
	rest := discordmocks.NewMockRest(t)

	svc := autorole.NewService(autorole.ServiceParams{
		Log:          zap.NewNop(),
		DiscordApi:   rest,
		AutoRoleRepo: repo,
	})

	return svc, repo, rest
}

func TestAddAutoRole(t *testing.T) {
	t.Run("saves a new auto role", func(t *testing.T) {
		svc, repo, _ := newService(t)

		repo.EXPECT().Find(mock.Anything, autorolerepo.FindParams{GuildID: testGuildID, RoleID: testRoleID}).Return(nil, nil)
		repo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(r *autorolerepo.AutoRole) bool {
			return r.GuildID == testGuildID && r.RoleID == testRoleID
		})).Return(nil)

		result, err := svc.AddAutoRole(t.Context(), autorole.AddAutoRoleParams{
			GuildID: testGuildID,
			RoleID:  testRoleID,
		})

		assert.NoError(t, err)
		assert.Equal(t, testGuildID, result.GuildID)
		assert.Equal(t, testRoleID, result.RoleID)
	})

	t.Run("rejects a duplicate role", func(t *testing.T) {
		svc, repo, _ := newService(t)

		repo.EXPECT().Find(mock.Anything, autorolerepo.FindParams{GuildID: testGuildID, RoleID: testRoleID}).Return([]autorolerepo.AutoRole{
			{GuildID: testGuildID, RoleID: testRoleID},
		}, nil)

		result, err := svc.AddAutoRole(t.Context(), autorole.AddAutoRoleParams{
			GuildID: testGuildID,
			RoleID:  testRoleID,
		})

		assert.Nil(t, result)
		assert.ErrorIs(t, err, autorole.ErrAutoRoleAlreadyExists)
	})

	t.Run("wraps the lookup error", func(t *testing.T) {
		svc, repo, _ := newService(t)

		repo.EXPECT().Find(mock.Anything, autorolerepo.FindParams{GuildID: testGuildID, RoleID: testRoleID}).Return(nil, errors.New("db down"))

		result, err := svc.AddAutoRole(t.Context(), autorole.AddAutoRoleParams{GuildID: testGuildID, RoleID: testRoleID})

		assert.Nil(t, result)
		assert.ErrorContains(t, err, "db down")
	})

	t.Run("wraps the save error", func(t *testing.T) {
		svc, repo, _ := newService(t)

		repo.EXPECT().Find(mock.Anything, autorolerepo.FindParams{GuildID: testGuildID, RoleID: testRoleID}).Return(nil, nil)
		repo.EXPECT().Save(mock.Anything, mock.Anything).Return(errors.New("write failed"))

		result, err := svc.AddAutoRole(t.Context(), autorole.AddAutoRoleParams{GuildID: testGuildID, RoleID: testRoleID})

		assert.Nil(t, result)
		assert.ErrorContains(t, err, "write failed")
	})
}

func TestRemoveAutoRole(t *testing.T) {
	t.Run("removes an existing role", func(t *testing.T) {
		svc, repo, _ := newService(t)

		repo.EXPECT().Delete(mock.Anything, autorolerepo.DeleteParams{GuildID: testGuildID, RoleID: testRoleID}).Return(int64(1), nil)

		err := svc.RemoveAutoRole(t.Context(), autorole.RemoveAutoRoleParams{GuildID: testGuildID, RoleID: testRoleID})

		assert.NoError(t, err)
	})

	t.Run("returns not found when nothing was deleted", func(t *testing.T) {
		svc, repo, _ := newService(t)

		repo.EXPECT().Delete(mock.Anything, autorolerepo.DeleteParams{GuildID: testGuildID, RoleID: testRoleID}).Return(int64(0), nil)

		err := svc.RemoveAutoRole(t.Context(), autorole.RemoveAutoRoleParams{GuildID: testGuildID, RoleID: testRoleID})

		assert.ErrorIs(t, err, autorole.ErrAutoRoleNotFound)
	})

	t.Run("wraps the repo error", func(t *testing.T) {
		svc, repo, _ := newService(t)

		repo.EXPECT().Delete(mock.Anything, autorolerepo.DeleteParams{GuildID: testGuildID, RoleID: testRoleID}).Return(int64(0), errors.New("db down"))

		err := svc.RemoveAutoRole(t.Context(), autorole.RemoveAutoRoleParams{GuildID: testGuildID, RoleID: testRoleID})

		assert.ErrorContains(t, err, "db down")
	})
}

func TestListAutoRoles(t *testing.T) {
	svc, repo, _ := newService(t)

	expected := []autorolerepo.AutoRole{{GuildID: testGuildID, RoleID: testRoleID}}
	repo.EXPECT().Find(mock.Anything, autorolerepo.FindParams{GuildID: testGuildID}).Return(expected, nil)

	result, err := svc.ListAutoRoles(t.Context(), autorole.ListAutoRolesParams{GuildID: testGuildID})

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestRoleDelete(t *testing.T) {
	svc, repo, _ := newService(t)

	repo.EXPECT().Delete(mock.Anything, autorolerepo.DeleteParams{GuildID: testGuildID, RoleID: testRoleID}).Return(int64(1), nil)

	err := svc.RoleDelete(t.Context(), autorole.RoleDeleteParams{GuildID: testGuildID, RoleID: testRoleID})

	assert.NoError(t, err)
}

func TestGuildMemberJoin(t *testing.T) {
	t.Run("adds every configured role to the member", func(t *testing.T) {
		svc, repo, rest := newService(t)

		repo.EXPECT().Find(mock.Anything, autorolerepo.FindParams{GuildID: testGuildID}).Return([]autorolerepo.AutoRole{
			{GuildID: testGuildID, RoleID: testRoleID},
			{GuildID: testGuildID, RoleID: testRoleID + 1},
		}, nil)
		rest.EXPECT().AddMemberRole(testGuildID, testUserID, testRoleID, mock.Anything).Return(nil)
		rest.EXPECT().AddMemberRole(testGuildID, testUserID, testRoleID+1, mock.Anything).Return(nil)

		err := svc.GuildMemberJoin(t.Context(), autorole.GuildMemberJoinParams{GuildID: testGuildID, UserID: testUserID})

		assert.NoError(t, err)
	})

	t.Run("wraps the lookup error", func(t *testing.T) {
		svc, repo, _ := newService(t)

		repo.EXPECT().Find(mock.Anything, autorolerepo.FindParams{GuildID: testGuildID}).Return(nil, errors.New("db down"))

		err := svc.GuildMemberJoin(t.Context(), autorole.GuildMemberJoinParams{GuildID: testGuildID, UserID: testUserID})

		assert.ErrorContains(t, err, "db down")
	})

	t.Run("continues assigning roles and joins the errors", func(t *testing.T) {
		svc, repo, rest := newService(t)

		repo.EXPECT().Find(mock.Anything, autorolerepo.FindParams{GuildID: testGuildID}).Return([]autorolerepo.AutoRole{
			{GuildID: testGuildID, RoleID: testRoleID},
			{GuildID: testGuildID, RoleID: testRoleID + 1},
		}, nil)
		rest.EXPECT().AddMemberRole(testGuildID, testUserID, testRoleID, mock.Anything).Return(errors.New("forbidden"))
		rest.EXPECT().AddMemberRole(testGuildID, testUserID, testRoleID+1, mock.Anything).Return(nil)

		err := svc.GuildMemberJoin(t.Context(), autorole.GuildMemberJoinParams{GuildID: testGuildID, UserID: testUserID})

		assert.ErrorContains(t, err, "forbidden")
	})
}
