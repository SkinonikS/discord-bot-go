package musicPlayer

import (
	"context"
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/v1/service/musicPlayer/player"
	"github.com/SkinonikS/discord-bot-go/internal/v1/service/musicPlayerSource"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/go-faster/errors"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type InteractionCommand struct {
	manager *Manager
	sources *musicPlayerSource.Registry
	log     *zap.SugaredLogger
}

type InteractionCommandParams struct {
	fx.In
	Manager *Manager
	Sources *musicPlayerSource.Registry
	Log     *zap.Logger
}

func NewInteractionCommand(p InteractionCommandParams) *InteractionCommand {
	return &InteractionCommand{
		manager: p.Manager,
		sources: p.Sources,
		log:     p.Log.Sugar(),
	}
}

func (c *InteractionCommand) Name() string {
	return "music"
}

func (c *InteractionCommand) Definition() disgodiscord.SlashCommandCreate {
	sources := lo.Map(c.sources.Names(), func(name string, _ int) disgodiscord.ApplicationCommandOptionChoiceString {
		return disgodiscord.ApplicationCommandOptionChoiceString{
			Name:  name,
			Value: name,
		}
	})

	sourceOpt := func(description string) disgodiscord.ApplicationCommandOptionString {
		return disgodiscord.ApplicationCommandOptionString{
			Name:        "source",
			Description: description,
			Required:    true,
			Choices:     sources,
		}
	}
	urlOpt := func(description string) disgodiscord.ApplicationCommandOptionString {
		return disgodiscord.ApplicationCommandOptionString{
			Name:        "url",
			Description: description,
			Required:    true,
		}
	}

	minIndex := 1
	return disgodiscord.SlashCommandCreate{
		Name:        c.Name(),
		Description: "Music player commands",
		Contexts:    []disgodiscord.InteractionContextType{disgodiscord.InteractionContextTypeGuild},
		Options: []disgodiscord.ApplicationCommandOption{
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "play",
				Description: "Clear the queue and play a track immediately",
				Options:     []disgodiscord.ApplicationCommandOption{sourceOpt("Audio source"), urlOpt("Track URL")},
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "add",
				Description: "Add a track to the end of the queue",
				Options:     []disgodiscord.ApplicationCommandOption{sourceOpt("Audio source"), urlOpt("Track URL")},
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "skip",
				Description: "Skip the current track",
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "stop",
				Description: "Stop playback and disconnect",
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "queue",
				Description: "Show the current queue",
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "pause",
				Description: "Pause playback",
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "resume",
				Description: "Resume playback",
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "clear",
				Description: "Clear queue",
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "playlist",
				Description: "Add all tracks from a playlist to the queue",
				Options:     []disgodiscord.ApplicationCommandOption{sourceOpt("Audio source"), urlOpt("Playlist URL")},
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "remove",
				Description: "Remove a track from the queue by position",
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionInt{
						Name:        "index",
						Description: "Position in queue (1-based)",
						Required:    true,
						MinValue:    &minIndex,
					},
				},
			},
		},
	}
}

func (c *InteractionCommand) Execute(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
	data := e.SlashCommandInteractionData()
	if data.SubCommandName == nil {
		return nil
	}

	switch *data.SubCommandName {
	case "play":
		return c.handlePlay(ctx, e, data, true)
	case "add":
		return c.handlePlay(ctx, e, data, false)
	case "queue":
		return c.handleQueue(ctx, e)
	case "skip":
		return c.handleSkip(ctx, e)
	case "stop":
		return c.handleStop(ctx, e)
	case "pause":
		return c.handlePause(ctx, e)
	case "resume":
		return c.handleResume(ctx, e)
	case "playlist":
		return c.handlePlaylist(ctx, e, data)
	case "remove":
		return c.handleRemove(ctx, e, data)
	case "clear":
		return c.handleClear(ctx, e)
	}

	return nil
}

func (c *InteractionCommand) handlePlay(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
	data disgodiscord.SlashCommandInteractionData,
	playNow bool,
) error {
	sourceName := data.Options["source"].String()
	url := data.Options["url"].String()

	if !c.sources.Has(sourceName) {
		return c.respondWithUnknownSource(ctx, e)
	}

	guildID := *e.GuildID()
	_, ok := e.Client().Caches.VoiceState(guildID, e.Member().User.ID)
	if !ok {
		return e.CreateMessage(disgodiscord.MessageCreate{
			Content: "You must be in a voice channel to use this command.",
			Flags:   disgodiscord.MessageFlagEphemeral,
		}, disgorest.WithCtx(ctx))
	}

	if err := e.DeferCreateMessage(true, disgorest.WithCtx(ctx)); err != nil {
		return err
	}

	client := e.Client()
	appID := e.ApplicationID()
	token := e.Token()

	go func() {
		track, err := c.manager.Play(context.Background(), PlayParams{
			GuildID:       guildID,
			UserID:        e.Member().User.ID,
			TextChannelID: e.Channel().ID(),
			URL:           url,
			SourceName:    sourceName,
			PlayNow:       playNow,
		})
		if err != nil {
			embeds := []disgodiscord.Embed{{
				Title:       "Playback error",
				Description: err.Error(),
				Color:       0xff0000,
			}}
			if _, err := client.Rest.UpdateInteractionResponse(appID, token, disgodiscord.MessageUpdate{
				Embeds: &embeds,
			}); err != nil {
				c.log.Warnw("failed to edit deferred response", zap.Error(err))
			}
			return
		}

		var title string
		if playNow {
			title = fmt.Sprintf("Playing now **%s** from **%s** (%s), requested by <@%s>", track.Title, track.Source, track.Duration, track.RequestedBy)
		} else {
			title = fmt.Sprintf("Added **%s** to queue from **%s** (%s), requested by <@%s>", track.Title, track.Source, track.Duration, track.RequestedBy)
		}

		if _, err := client.Rest.UpdateInteractionResponse(appID, token, disgodiscord.MessageUpdate{
			Content: &title,
		}); err != nil {
			c.log.Warnw("failed to edit deferred response", zap.Error(err))
		}
	}()

	return nil
}

func (c *InteractionCommand) handleQueue(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
	queue, err := c.manager.Queue(*e.GuildID())
	if err != nil {
		if errors.Is(err, ErrNotInitialized) {
			return c.respondNotInVoiceChannel(ctx, e)
		}
		return err
	}

	if queue.Len() == 0 {
		return e.CreateMessage(disgodiscord.MessageCreate{
			Content: "The queue is empty.",
			Flags:   disgodiscord.MessageFlagEphemeral,
		}, disgorest.WithCtx(ctx))
	}

	const maxFields = 25
	tracks := queue.List()
	total := len(tracks)
	if total > maxFields {
		tracks = tracks[:maxFields]
	}

	var fields []disgodiscord.EmbedField
	for i, track := range tracks {
		var name string
		if i == 0 {
			name = fmt.Sprintf("Now playing - %s", track.Title)
		} else {
			name = fmt.Sprintf("#%d - %s", i, track.Title)
		}
		fields = append(fields, disgodiscord.EmbedField{
			Name:  name,
			Value: fmt.Sprintf("%s | %s | Requested by <@%s>", track.Source, track.Duration, track.RequestedBy),
		})
	}

	var footer *disgodiscord.EmbedFooter
	if total > maxFields {
		footer = &disgodiscord.EmbedFooter{
			Text: fmt.Sprintf("Showing first %d of %d tracks", maxFields, total),
		}
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Embeds: []disgodiscord.Embed{
			{
				Title:  fmt.Sprintf("Queue (%d tracks)", total),
				Color:  0x00ff00,
				Fields: fields,
				Footer: footer,
			},
		},
		Flags: disgodiscord.MessageFlagEphemeral,
	}, disgorest.WithCtx(ctx))
}

func (c *InteractionCommand) handlePause(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
	if err := c.manager.Pause(*e.GuildID()); err != nil {
		if errors.Is(err, ErrNotInitialized) {
			return c.respondNotInVoiceChannel(ctx, e)
		}
		if errors.Is(err, ErrNotPlaying) {
			return e.CreateMessage(disgodiscord.MessageCreate{
				Content: "Nothing is currently playing.",
				Flags:   disgodiscord.MessageFlagEphemeral,
			}, disgorest.WithCtx(ctx))
		}
		return err
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Content: "Playback paused.",
		Flags:   disgodiscord.MessageFlagEphemeral,
	}, disgorest.WithCtx(ctx))
}

func (c *InteractionCommand) handleResume(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
	if err := c.manager.Resume(*e.GuildID()); err != nil {
		if errors.Is(err, ErrNotInitialized) {
			return c.respondNotInVoiceChannel(ctx, e)
		}
		if errors.Is(err, ErrNotPaused) {
			return e.CreateMessage(disgodiscord.MessageCreate{
				Content: "Playback is not paused.",
				Flags:   disgodiscord.MessageFlagEphemeral,
			}, disgorest.WithCtx(ctx))
		}
		return err
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Content: "Playback resumed.",
		Flags:   disgodiscord.MessageFlagEphemeral,
	}, disgorest.WithCtx(ctx))
}

func (c *InteractionCommand) handleStop(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
	if err := c.manager.Stop(*e.GuildID()); err != nil {
		if errors.Is(err, ErrNotInitialized) {
			return c.respondNotInVoiceChannel(ctx, e)
		}
		return err
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Content: "Playback stopped. Queue cleared.",
		Flags:   disgodiscord.MessageFlagEphemeral,
	}, disgorest.WithCtx(ctx))
}

func (c *InteractionCommand) handleSkip(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
	if err := c.manager.Skip(*e.GuildID()); err != nil {
		if errors.Is(err, ErrNotInitialized) {
			return c.respondNotInVoiceChannel(ctx, e)
		}
		return err
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Content: "Skipped current track.",
		Flags:   disgodiscord.MessageFlagEphemeral,
	}, disgorest.WithCtx(ctx))
}

func (c *InteractionCommand) handleRemove(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
	data disgodiscord.SlashCommandInteractionData,
) error {
	index := data.Options["index"].Int()

	track, err := c.manager.Remove(*e.GuildID(), index)
	if err != nil {
		if errors.Is(err, ErrNotInitialized) {
			return c.respondNotInVoiceChannel(ctx, e)
		}
		if errors.Is(err, player.ErrInvalidIndex) {
			return e.CreateMessage(disgodiscord.MessageCreate{
				Content: "Invalid queue position.",
				Flags:   disgodiscord.MessageFlagEphemeral,
			}, disgorest.WithCtx(ctx))
		}
		return err
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Content: fmt.Sprintf("Removed **%s** from the queue.", track.Title),
		Flags:   disgodiscord.MessageFlagEphemeral,
	}, disgorest.WithCtx(ctx))
}

func (c *InteractionCommand) handlePlaylist(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
	data disgodiscord.SlashCommandInteractionData,
) error {
	sourceName := data.Options["source"].String()
	url := data.Options["url"].String()

	if !c.sources.Has(sourceName) {
		return c.respondWithUnknownSource(ctx, e)
	}

	if err := e.DeferCreateMessage(true, disgorest.WithCtx(ctx)); err != nil {
		return err
	}

	guildID := *e.GuildID()
	client := e.Client()
	appID := e.ApplicationID()
	token := e.Token()

	go func() {
		tracks, err := c.manager.AddPlaylist(context.Background(), PlaylistParams{
			GuildID:       guildID,
			UserID:        e.Member().User.ID,
			TextChannelID: e.Channel().ID(),
			URL:           url,
			SourceName:    sourceName,
		})
		if err != nil {
			embeds := []disgodiscord.Embed{{
				Title:       "Playlist error",
				Description: err.Error(),
				Color:       0xff0000,
			}}
			if _, err := client.Rest.UpdateInteractionResponse(appID, token, disgodiscord.MessageUpdate{
				Embeds: &embeds,
			}); err != nil {
				c.log.Warnw("failed to edit deferred response", zap.Error(err))
			}
			return
		}

		content := fmt.Sprintf("Added **%d** tracks from **%s** playlist to the queue.", len(tracks), sourceName)
		if _, err := client.Rest.UpdateInteractionResponse(appID, token, disgodiscord.MessageUpdate{
			Content: &content,
		}); err != nil {
			c.log.Warnw("failed to edit deferred response", zap.Error(err))
		}
	}()

	return nil
}

func (c *InteractionCommand) handleClear(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
	if err := c.manager.EraseQueue(*e.GuildID()); err != nil {
		if errors.Is(err, ErrNotInitialized) {
			return c.respondNotInVoiceChannel(ctx, e)
		}
		return err
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Content: "Queue cleared.",
		Flags:   disgodiscord.MessageFlagEphemeral,
	}, disgorest.WithCtx(ctx))
}

func (c *InteractionCommand) respondWithUnknownSource(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
	return e.CreateMessage(disgodiscord.MessageCreate{
		Content: "Unknown source. Please choose from the list of available sources.",
		Flags:   disgodiscord.MessageFlagEphemeral,
	}, disgorest.WithCtx(ctx))
}

func (c *InteractionCommand) respondNotInVoiceChannel(
	ctx context.Context,
	e *disgoevents.ApplicationCommandInteractionCreate,
) error {
	return e.CreateMessage(disgodiscord.MessageCreate{
		Content: "Music player is not initialized in this guild. Please join a voice channel and use the `/music play` command to initialize it.",
		Flags:   disgodiscord.MessageFlagEphemeral,
	}, disgorest.WithCtx(ctx))
}
