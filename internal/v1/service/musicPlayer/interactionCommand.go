package musicPlayer

import (
	"context"
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/v1/service/musicPlayer/player"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/musicPlayerSource"
	"github.com/SkinonikS/discord-bot-go/internal/v1/translator"
	"github.com/SkinonikS/discord-bot-go/internal/v1/util"
	"github.com/bwmarrin/discordgo"
	"github.com/go-faster/errors"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type InteractionCommand struct {
	t       *translator.Translator
	manager *Manager
	sources *musicPlayerSource.Registry
	log     *zap.SugaredLogger
}

type InteractionCommandParams struct {
	fx.In
	T       *translator.Translator
	Manager *Manager
	Sources *musicPlayerSource.Registry
	Log     *zap.Logger
}

func NewInteractionCommand(p InteractionCommandParams) *InteractionCommand {
	return &InteractionCommand{
		t:       p.T,
		manager: p.Manager,
		sources: p.Sources,
		log:     p.Log.Sugar(),
	}
}

func (c *InteractionCommand) Name() string {
	return "music"
}

func (c *InteractionCommand) ForOwnerOnly() bool {
	return false
}

func (c *InteractionCommand) Definition() *discordgo.ApplicationCommand {
	sources := lo.Map(c.sources.Names(), func(name string, _ int) *discordgo.ApplicationCommandOptionChoice {
		return &discordgo.ApplicationCommandOptionChoice{
			Name:  name,
			Value: name,
		}
	})

	return &discordgo.ApplicationCommand{

		Name:        c.Name(),
		Description: "Music player commands",
		Contexts:    &[]discordgo.InteractionContextType{discordgo.InteractionContextGuild},
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "play",
				Description: "Clear the queue and play a track immediately",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "source",
						Description: "Audio source",
						Required:    true,
						Choices:     sources,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "url",
						Description: "Track URL",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add a track to the end of the queue",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "source",
						Description: "Audio source",
						Required:    true,
						Choices:     sources,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "url",
						Description: "Track URL",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "skip",
				Description: "Skip the current track",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "stop",
				Description: "Stop playback and disconnect",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "queue",
				Description: "Show the current queue",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "pause",
				Description: "Pause playback",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "resume",
				Description: "Resume playback",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a track from the queue by position",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "index",
						Description: "Position in queue (1-based)",
						Required:    true,
						MinValue:    util.ToPtr(float64(1)),
					},
				},
			},
		},
	}
}

func (c *InteractionCommand) Execute(
	ctx context.Context,
	s *discordgo.Session,
	e *discordgo.InteractionCreate,
) error {
	if e.Type != discordgo.InteractionApplicationCommand {
		return nil
	}

	data := e.ApplicationCommandData()
	if len(data.Options) == 0 {
		return nil
	}

	sub := data.Options[0]

	switch sub.Name {
	case "play":
		return c.handlePlay(ctx, s, e, sub, true)
	case "add":
		return c.handlePlay(ctx, s, e, sub, false)
	case "queue":
		return c.handleQueue(ctx, s, e)
	case "skip":
		return c.handleSkip(ctx, s, e)
	case "stop":
		return c.handleStop(ctx, s, e)
	case "pause":
		return c.handlePause(ctx, s, e)
	case "resume":
		return c.handleResume(ctx, s, e)
	case "remove":
		return c.handleRemove(ctx, s, e)
	}

	return nil
}

func (c *InteractionCommand) handlePlay(
	ctx context.Context,
	s *discordgo.Session,
	e *discordgo.InteractionCreate,
	sub *discordgo.ApplicationCommandInteractionDataOption,
	playNow bool,
) error {
	sourceName := sub.Options[0].StringValue()
	url := sub.Options[1].StringValue()

	if !c.sources.Has(sourceName) {
		return c.respondWithUnknownSource(ctx, s, e)
	}

	if err := s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}, discordgo.WithContext(ctx)); err != nil {
		return err
	}

	go func() {
		track, err := c.manager.Play(context.Background(), PlayParams{
			GuildID:       e.GuildID,
			UserID:        e.Member.User.ID,
			TextChannelID: e.ChannelID,
			URL:           url,
			SourceName:    sourceName,
			PlayNow:       playNow,
		})
		if err != nil {
			if _, err := s.InteractionResponseEdit(e.Interaction, &discordgo.WebhookEdit{
				Content: nil,
				Embeds: &[]*discordgo.MessageEmbed{
					{
						Title:       "Playback error",
						Description: err.Error(),
						Color:       0xff0000,
					},
				},
			}); err != nil {
				c.log.Warnw("failed to edit deferred response", zap.Error(err))
			}
			return
		}

		title := fmt.Sprintf("Playing now **%s** from **%s** (%s), requested by <@%s>", track.Title, track.Source, track.Duration, track.RequestedBy)
		if !playNow {
			title = fmt.Sprintf("Added **%s** to queue from **%s** (%s), requested by <@%s>", track.Title, track.Source, track.Duration, track.RequestedBy)
		}

		if _, err := s.InteractionResponseEdit(e.Interaction, &discordgo.WebhookEdit{
			Content: util.ToPtr(title),
		}); err != nil {
			c.log.Warnw("failed to edit deferred response", zap.Error(err))
		}
	}()

	return nil
}

func (c *InteractionCommand) handleQueue(
	ctx context.Context,
	s *discordgo.Session,
	e *discordgo.InteractionCreate,
) error {
	queue, err := c.manager.Queue(e.GuildID)
	if err != nil {
		if errors.Is(err, ErrNotInitialized) {
			return c.respondNotInVoiceChannel(ctx, s, e)
		}
		return err
	}

	if queue.Len() == 0 {
		return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "The queue is empty.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}, discordgo.WithContext(ctx))
	}

	var fields []*discordgo.MessageEmbedField
	for i, track := range queue.List() {
		var name string
		if i == 0 {
			name = fmt.Sprintf("Now playing - %s", track.Title)
		} else {
			name = fmt.Sprintf("#%d - %s", i, track.Title)
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  name,
			Value: fmt.Sprintf("%s | %s | Requested by <@%s>", track.Source, track.Duration, track.RequestedBy),
		})
	}

	return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:  "Queue",
					Color:  0x00ff00,
					Fields: fields,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}, discordgo.WithContext(ctx))
}

func (c *InteractionCommand) handlePause(ctx context.Context, s *discordgo.Session, e *discordgo.InteractionCreate) error {
	if err := c.manager.Pause(e.GuildID); err != nil {
		if errors.Is(err, ErrNotInitialized) {
			return c.respondNotInVoiceChannel(ctx, s, e)
		}
		if errors.Is(err, ErrNotPlaying) {
			return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Nothing is currently playing.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}, discordgo.WithContext(ctx))
		}
		return err
	}

	return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Playback paused.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}, discordgo.WithContext(ctx))
}

func (c *InteractionCommand) handleResume(
	ctx context.Context,
	s *discordgo.Session,
	e *discordgo.InteractionCreate,
) error {
	if err := c.manager.Resume(e.GuildID); err != nil {
		if errors.Is(err, ErrNotInitialized) {
			return c.respondNotInVoiceChannel(ctx, s, e)
		}
		if errors.Is(err, ErrNotPaused) {
			return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Playback is not paused.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}, discordgo.WithContext(ctx))
		}
		return err
	}

	return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Playback resumed.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}, discordgo.WithContext(ctx))
}

func (c *InteractionCommand) handleStop(
	ctx context.Context,
	s *discordgo.Session,
	e *discordgo.InteractionCreate,
) error {
	if err := c.manager.Stop(e.GuildID); err != nil {
		if errors.Is(err, ErrNotInitialized) {
			return c.respondNotInVoiceChannel(ctx, s, e)
		}
		return err
	}

	return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Playback stopped. Queue cleared.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (c *InteractionCommand) handleSkip(
	ctx context.Context,
	s *discordgo.Session,
	e *discordgo.InteractionCreate,
) error {
	if err := c.manager.Skip(e.GuildID); err != nil {
		if errors.Is(err, ErrNotInitialized) {
			return c.respondNotInVoiceChannel(ctx, s, e)
		}
		return err
	}

	return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Skipped current track.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (c *InteractionCommand) handleRemove(
	ctx context.Context,
	s *discordgo.Session,
	e *discordgo.InteractionCreate,
) error {
	index := int(e.ApplicationCommandData().Options[0].Options[0].IntValue())

	track, err := c.manager.Remove(e.GuildID, index)
	if err != nil {
		if errors.Is(err, ErrNotInitialized) {
			return c.respondNotInVoiceChannel(ctx, s, e)
		}
		if errors.Is(err, player.ErrInvalidIndex) {
			return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Invalid queue position.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			}, discordgo.WithContext(ctx))
		}
		return err
	}

	return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Removed **%s** from the queue.", track.Title),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}, discordgo.WithContext(ctx))
}

func (c *InteractionCommand) respondWithUnknownSource(
	ctx context.Context,
	s *discordgo.Session,
	e *discordgo.InteractionCreate,
) (err error) {
	return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Unknown source. Please choose from the list of available sources.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}, discordgo.WithContext(ctx))
}

func (c *InteractionCommand) respondNotInVoiceChannel(
	ctx context.Context,
	s *discordgo.Session,
	e *discordgo.InteractionCreate,
) (err error) {
	return s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Music player is not initialized in this guild. Please join a voice channel and use the `/music play` command to initialize it.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}, discordgo.WithContext(ctx))
}
