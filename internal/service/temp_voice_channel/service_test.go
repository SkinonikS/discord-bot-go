package tempvoicechannel_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"slices"
	"testing"
	"uuid"

	discordmocks "github.com/SkinonikS/discord-bot-go/internal/infra/discord/mock"
	translatormocks "github.com/SkinonikS/discord-bot-go/internal/infra/translator/mock"
	tempvoicechannelrepo "github.com/SkinonikS/discord-bot-go/internal/service/repository/temp_voice_channel"
	tempvoicechannelmocks "github.com/SkinonikS/discord-bot-go/internal/service/repository/temp_voice_channel/mock"
	tempvoicechannelstate "github.com/SkinonikS/discord-bot-go/internal/service/repository/temp_voice_channel_state"
	tempvoicechannelstatemocks "github.com/SkinonikS/discord-bot-go/internal/service/repository/temp_voice_channel_state/mock"
	"github.com/SkinonikS/discord-bot-go/internal/service/temp_voice_channel"
	disgodiscord "github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	testGuildID       = snowflake.ID(1)
	testRootChannelID = snowflake.ID(2)
	testParentID      = snowflake.ID(3)
	testOwnerID       = snowflake.ID(4)
	testChannelID     = snowflake.ID(5)
)

type mocks struct {
	translator       *translatormocks.MockTranslator
	rest             *discordmocks.MockRest
	cache            *discordmocks.MockCaches
	channelRepo      *tempvoicechannelmocks.MockRepo
	channelStateRepo *tempvoicechannelstatemocks.MockRepo
}

func newService(t *testing.T) (tempvoicechannel.Service, mocks) {
	t.Helper()

	m := mocks{
		translator:       translatormocks.NewMockTranslator(t),
		rest:             discordmocks.NewMockRest(t),
		cache:            discordmocks.NewMockCaches(t),
		channelRepo:      tempvoicechannelmocks.NewMockRepo(t),
		channelStateRepo: tempvoicechannelstatemocks.NewMockRepo(t),
	}

	svc := tempvoicechannel.NewService(tempvoicechannel.ServiceParams{
		T:                m.translator,
		DiscordApi:       m.rest,
		BotCache:         m.cache,
		ChannelRepo:      m.channelRepo,
		ChannelStateRepo: m.channelStateRepo,
		Log:              zap.NewNop(),
	})

	return svc, m
}

func newGuildVoiceChannel(t *testing.T, id snowflake.ID) *disgodiscord.GuildVoiceChannel {
	t.Helper()

	ch := &disgodiscord.GuildVoiceChannel{}
	require.NoError(t, json.Unmarshal([]byte(fmt.Sprintf(`{"id":"%d"}`, id)), ch))
	return ch
}

func voiceStates(states ...disgodiscord.VoiceState) func(snowflake.ID) iter.Seq[disgodiscord.VoiceState] {
	return func(snowflake.ID) iter.Seq[disgodiscord.VoiceState] {
		return slices.Values(states)
	}
}

func withStateTransaction(repo *tempvoicechannelstatemocks.MockRepo) {
	repo.EXPECT().Transaction(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(tx tempvoicechannelstate.Repo) error) error {
		return fn(repo)
	})
}

func TestCreateSetupChannel(t *testing.T) {
	t.Run("saves the setup channel", func(t *testing.T) {
		svc, m := newService(t)

		m.channelRepo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(c *tempvoicechannelrepo.TempVoiceChannel) bool {
			return c.GuildID == testGuildID && c.RootChannelID == testRootChannelID && c.ParentID == testParentID
		})).Return(nil)

		result, err := svc.CreateSetupChannel(t.Context(), tempvoicechannel.SetupChannel{
			GuildID:          testGuildID,
			RootChannelID:    testRootChannelID,
			ParentCategoryID: testParentID,
		})

		assert.NoError(t, err)
		assert.Equal(t, testRootChannelID, result.RootChannelID)
	})

	t.Run("wraps the save error", func(t *testing.T) {
		svc, m := newService(t)

		m.channelRepo.EXPECT().Save(mock.Anything, mock.Anything).Return(errors.New("write failed"))

		_, err := svc.CreateSetupChannel(t.Context(), tempvoicechannel.SetupChannel{})

		assert.ErrorContains(t, err, "write failed")
	})
}

func TestDeleteSetupChannel(t *testing.T) {
	t.Run("deletes every matching setup channel", func(t *testing.T) {
		svc, m := newService(t)

		id := uuid.New()
		m.channelRepo.EXPECT().Find(mock.Anything, tempvoicechannelrepo.FindParams{
			GuildID:       testGuildID,
			RootChannelID: testRootChannelID,
		}).Return([]tempvoicechannelrepo.TempVoiceChannel{{ID: id}}, nil)
		m.channelRepo.EXPECT().DeleteManyByIDs(mock.Anything, []uuid.UUID{id}).Return(int64(1), nil)

		err := svc.DeleteSetupChannel(t.Context(), tempvoicechannel.DeleteSetupChannel{
			GuildID:       testGuildID,
			RootChannelID: testRootChannelID,
		})

		assert.NoError(t, err)
	})

	t.Run("returns ErrSetupChannelNotFound when nothing matches", func(t *testing.T) {
		svc, m := newService(t)

		m.channelRepo.EXPECT().Find(mock.Anything, mock.Anything).Return(nil, nil)

		err := svc.DeleteSetupChannel(t.Context(), tempvoicechannel.DeleteSetupChannel{})

		assert.ErrorIs(t, err, tempvoicechannel.ErrSetupChannelNotFound)
	})

	t.Run("wraps the lookup error", func(t *testing.T) {
		svc, m := newService(t)

		m.channelRepo.EXPECT().Find(mock.Anything, mock.Anything).Return(nil, errors.New("db down"))

		err := svc.DeleteSetupChannel(t.Context(), tempvoicechannel.DeleteSetupChannel{})

		assert.ErrorContains(t, err, "db down")
	})
}

func TestLeaveChannel(t *testing.T) {
	t.Run("returns ErrNotTempChannel when no state is found", func(t *testing.T) {
		svc, m := newService(t)

		m.channelStateRepo.EXPECT().Find(mock.Anything, tempvoicechannelstate.FindParams{
			GuildID:   testGuildID,
			ChannelID: testChannelID,
		}).Return(nil, nil)

		err := svc.LeaveChannel(t.Context(), tempvoicechannel.LeaveChannel{GuildID: testGuildID, ChannelID: testChannelID})

		assert.ErrorIs(t, err, tempvoicechannel.ErrNotTempChannel)
	})

	t.Run("wraps the lookup error", func(t *testing.T) {
		svc, m := newService(t)

		m.channelStateRepo.EXPECT().Find(mock.Anything, mock.Anything).Return(nil, errors.New("db down"))

		err := svc.LeaveChannel(t.Context(), tempvoicechannel.LeaveChannel{GuildID: testGuildID, ChannelID: testChannelID})

		assert.ErrorContains(t, err, "db down")
	})

	t.Run("deletes the channel once the last member leaves", func(t *testing.T) {
		svc, m := newService(t)

		stateID := uuid.New()
		m.channelStateRepo.EXPECT().Find(mock.Anything, mock.Anything).Return([]tempvoicechannelstate.TempVoiceChannelState{{ID: stateID}}, nil)
		m.cache.EXPECT().VoiceStates(testGuildID).RunAndReturn(voiceStates())
		m.rest.EXPECT().DeleteChannel(testChannelID, mock.Anything).Return(nil)
		withStateTransaction(m.channelStateRepo)
		m.channelStateRepo.EXPECT().DeleteManyByIDs(mock.Anything, []uuid.UUID{stateID}).Return(int64(1), nil)

		err := svc.LeaveChannel(t.Context(), tempvoicechannel.LeaveChannel{GuildID: testGuildID, ChannelID: testChannelID})

		assert.NoError(t, err)
	})

	t.Run("keeps the channel while members remain", func(t *testing.T) {
		svc, m := newService(t)

		m.channelStateRepo.EXPECT().Find(mock.Anything, mock.Anything).Return([]tempvoicechannelstate.TempVoiceChannelState{{ID: uuid.New()}}, nil)
		m.cache.EXPECT().VoiceStates(testGuildID).RunAndReturn(voiceStates(disgodiscord.VoiceState{ChannelID: new(testChannelID)}))

		err := svc.LeaveChannel(t.Context(), tempvoicechannel.LeaveChannel{GuildID: testGuildID, ChannelID: testChannelID})

		assert.NoError(t, err)
		m.rest.AssertNotCalled(t, "DeleteChannel", mock.Anything, mock.Anything)
	})
}

func TestJoinChannel(t *testing.T) {
	t.Run("returns ErrNotSetupChannel when no setup channel is found", func(t *testing.T) {
		svc, m := newService(t)

		m.channelRepo.EXPECT().Find(mock.Anything, tempvoicechannelrepo.FindParams{
			GuildID:       testGuildID,
			RootChannelID: testRootChannelID,
		}).Return(nil, nil)

		_, err := svc.JoinChannel(t.Context(), tempvoicechannel.JoinChannel{GuildID: testGuildID, SetupChannelID: testRootChannelID})

		assert.ErrorIs(t, err, tempvoicechannel.ErrNotSetupChannel)
	})

	t.Run("wraps the lookup error", func(t *testing.T) {
		svc, m := newService(t)

		m.channelRepo.EXPECT().Find(mock.Anything, mock.Anything).Return(nil, errors.New("db down"))

		_, err := svc.JoinChannel(t.Context(), tempvoicechannel.JoinChannel{GuildID: testGuildID, SetupChannelID: testRootChannelID})

		assert.ErrorContains(t, err, "db down")
	})

	t.Run("creates the voice channel and moves the owner into it", func(t *testing.T) {
		svc, m := newService(t)

		m.channelRepo.EXPECT().Find(mock.Anything, mock.Anything).Return([]tempvoicechannelrepo.TempVoiceChannel{
			{GuildID: testGuildID, RootChannelID: testRootChannelID, ParentID: testParentID},
		}, nil)
		m.cache.EXPECT().Guild(testGuildID).Return(disgodiscord.Guild{}, false)
		m.translator.EXPECT().Localize(mock.Anything, mock.Anything).Return("owner's channel")

		created := newGuildVoiceChannel(t, testChannelID)
		m.rest.EXPECT().CreateGuildChannel(testGuildID, mock.MatchedBy(func(c disgodiscord.GuildVoiceChannelCreate) bool {
			return c.Name == "owner's channel" && c.ParentID == testParentID
		}), mock.Anything).Return(created, nil)
		m.rest.EXPECT().UpdateMember(testGuildID, testOwnerID, mock.MatchedBy(func(u disgodiscord.MemberUpdate) bool {
			return u.ChannelID != nil && *u.ChannelID == testChannelID
		}), mock.Anything).Return(nil, nil)

		withStateTransaction(m.channelStateRepo)
		m.channelStateRepo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(s *tempvoicechannelstate.TempVoiceChannelState) bool {
			return s.ChannelID == testChannelID && s.GuildID == testGuildID
		})).Return(nil)

		result, err := svc.JoinChannel(t.Context(), tempvoicechannel.JoinChannel{
			GuildID:        testGuildID,
			SetupChannelID: testRootChannelID,
			OwnerID:        testOwnerID,
			OwnerName:      "owner",
		})

		require.NoError(t, err)
		assert.Equal(t, testChannelID, result.ID())
	})

	t.Run("does not try to delete anything when the channel could never be created", func(t *testing.T) {
		svc, m := newService(t)

		m.channelRepo.EXPECT().Find(mock.Anything, mock.Anything).Return([]tempvoicechannelrepo.TempVoiceChannel{
			{GuildID: testGuildID, RootChannelID: testRootChannelID, ParentID: testParentID},
		}, nil)
		m.cache.EXPECT().Guild(testGuildID).Return(disgodiscord.Guild{}, false)
		m.translator.EXPECT().Localize(mock.Anything, mock.Anything).Return("owner's channel")
		m.rest.EXPECT().CreateGuildChannel(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("rate limited"))
		withStateTransaction(m.channelStateRepo)

		_, err := svc.JoinChannel(t.Context(), tempvoicechannel.JoinChannel{
			GuildID:        testGuildID,
			SetupChannelID: testRootChannelID,
			OwnerID:        testOwnerID,
		})

		assert.ErrorContains(t, err, "rate limited")
		m.rest.AssertNotCalled(t, "DeleteChannel", mock.Anything, mock.Anything)
	})

	t.Run("deletes the created channel when moving the owner fails", func(t *testing.T) {
		svc, m := newService(t)

		m.channelRepo.EXPECT().Find(mock.Anything, mock.Anything).Return([]tempvoicechannelrepo.TempVoiceChannel{
			{GuildID: testGuildID, RootChannelID: testRootChannelID, ParentID: testParentID},
		}, nil)
		m.cache.EXPECT().Guild(testGuildID).Return(disgodiscord.Guild{}, false)
		m.translator.EXPECT().Localize(mock.Anything, mock.Anything).Return("owner's channel")

		created := newGuildVoiceChannel(t, testChannelID)
		m.rest.EXPECT().CreateGuildChannel(mock.Anything, mock.Anything, mock.Anything).Return(created, nil)
		m.rest.EXPECT().UpdateMember(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("cannot move member"))
		withStateTransaction(m.channelStateRepo)
		m.rest.EXPECT().DeleteChannel(testChannelID, mock.Anything).Return(nil)

		_, err := svc.JoinChannel(t.Context(), tempvoicechannel.JoinChannel{
			GuildID:        testGuildID,
			SetupChannelID: testRootChannelID,
			OwnerID:        testOwnerID,
		})

		assert.ErrorContains(t, err, "cannot move member")
	})
}
