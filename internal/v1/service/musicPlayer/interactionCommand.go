package musicPlayer

import (
	"context"
	"fmt"
	"regexp"

	disgobot "github.com/disgoorg/disgo/bot"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/lavaqueue-plugin"
	"go.uber.org/fx"
)

type InteractionCommand struct {
	lavaLinkClient disgolink.Client
	discordClient  *disgobot.Client
	urlPattern     *regexp.Regexp
	searchPattern  *regexp.Regexp
}

type InteractionCommandParams struct {
	fx.In
	LavaLinkClient disgolink.Client
	DiscordClient  *disgobot.Client
}

func NewInteractionCommand(p InteractionCommandParams) *InteractionCommand {
	return &InteractionCommand{
		lavaLinkClient: p.LavaLinkClient,
		discordClient:  p.DiscordClient,
		urlPattern:     regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?"),
		searchPattern:  regexp.MustCompile(`^(.{2})search:(.+)`),
	}
}

func (c *InteractionCommand) Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()

	switch *data.SubCommandName {
	case "play":
		return c.handlePlay(ctx, e)
	case "queue":
		return c.handleQueue(ctx, e)
	case "skip":
		return c.handleSkip(ctx, e)
	case "stop":
		return c.handleStop(ctx, e)
	}

	return fmt.Errorf("unknown subcommand: %s", *data.SubCommandName)
}

func (c *InteractionCommand) Definition() disgodiscord.SlashCommandCreate {
	return disgodiscord.SlashCommandCreate{
		Name:        c.Name(),
		Description: "Music player commands",
		Options: []disgodiscord.ApplicationCommandOption{
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "queue",
				Description: "Show the current queue",
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "stop",
				Description: "Stop playback and disconnect",
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "skip",
				Description: "Skips the current song",
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionInt{
						Name:        "count",
						Description: "The number of tracks to skip",
						Required:    false,
					},
				},
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:        "play",
				Description: "Play a track immediately",
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionString{
						Name:        "identifier",
						Description: "Track search query or url",
						Required:    true,
					},
					disgodiscord.ApplicationCommandOptionString{
						Name:        "source",
						Description: "The source to search on",
						Required:    false,
						Choices: []disgodiscord.ApplicationCommandOptionChoiceString{
							{
								Name:  "YouTube",
								Value: string(lavalink.SearchTypeYouTube),
							},
							{
								Name:  "YouTube Music",
								Value: string(lavalink.SearchTypeYouTubeMusic),
							},
						},
					},
				},
			},
		},
	}
}

func (c *InteractionCommand) Name() string {
	return "music"
}

func (c *InteractionCommand) handleSkip(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	player := c.lavaLinkClient.ExistingPlayer(*e.GuildID())
	if player == nil {
		return e.CreateMessage(disgodiscord.MessageCreate{
			Flags:   disgodiscord.MessageFlagEphemeral,
			Content: "No active player found.",
		}, disgorest.WithCtx(ctx))
	}

	if err := e.DeferCreateMessage(true, disgorest.WithCtx(ctx)); err != nil {
		return err
	}

	data := e.SlashCommandInteractionData()

	count := data.Int("count")
	if count < 1 {
		count = 1
	}

	track, err := lavaqueue.QueueNextTrack(context.Background(), player.Node(), *e.GuildID(), count)
	if err != nil {
		_, err = e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
			Content: new("error while skipping track"),
		})
		return err
	}

	if track == nil {
		_, err = e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
			Content: new("no tracks in queue"),
		})
		return err
	}

	_, err = e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
		Content: new(fmt.Sprintf("Playing: %s", track.Info.Title)),
	})
	return err
}

func (c *InteractionCommand) handleStop(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	player := c.lavaLinkClient.ExistingPlayer(*e.GuildID())
	if player == nil {
		return e.CreateMessage(disgodiscord.MessageCreate{
			Flags:   disgodiscord.MessageFlagEphemeral,
			Content: "No active player found.",
		}, disgorest.WithCtx(ctx))
	}

	if err := e.Client().UpdateVoiceState(ctx, *e.GuildID(), nil, false, false); err != nil {
		return fmt.Errorf("error while updating voice state: %w", err)
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags:   disgodiscord.MessageFlagEphemeral,
		Content: "Playback stopped.",
	})
}

func (c *InteractionCommand) handleQueue(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	queue, err := lavaqueue.GetQueue(ctx, c.lavaLinkClient.BestNode(), *e.GuildID())
	if err != nil {
		return err
	}

	if len(queue.Tracks) == 0 {
		_, err = c.discordClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
			Content: new("No tracks in queue."),
		}, disgorest.WithCtx(ctx))
		return err
	}

	var tracks string
	for i, track := range queue.Tracks {
		tracks += fmt.Sprintf("%d. [`%s`](<%s>)\n", i+1, track.Info.Title, *track.Info.URI)
	}

	var content string
	if queue.Type == "" {
		content = fmt.Sprintf("Queue:\n%s", tracks)
	} else {
		content = fmt.Sprintf("Queue `%s`:\n%s", queue.Type, tracks)
	}

	_, err = c.discordClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
		Content: new(content),
	}, disgorest.WithCtx(ctx))
	return err
}

func (c *InteractionCommand) handlePlay(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()

	identifier := data.String("identifier")
	if source, ok := data.OptString("source"); ok {
		identifier = lavalink.SearchType(source).Apply(identifier)
	} else if !c.urlPattern.MatchString(identifier) && !c.searchPattern.MatchString(identifier) {
		identifier = lavalink.SearchTypeYouTube.Apply(identifier)
	}

	voiceState, ok := c.discordClient.Caches.VoiceState(*e.GuildID(), e.User().ID)
	if !ok {
		return e.CreateMessage(disgodiscord.MessageCreate{
			Flags:   disgodiscord.MessageFlagEphemeral,
			Content: "You need to be in a voice channel to use this command.",
		})
	}

	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	if err := c.discordClient.UpdateVoiceState(ctx, *e.GuildID(), voiceState.ChannelID, false, true); err != nil {
		return err
	}

	node := c.lavaLinkClient.BestNode()

	var tracksToQueue []lavaqueue.QueueTrack
	node.LoadTracksHandler(ctx, identifier, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			_, _ = c.discordClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new(fmt.Sprintf("Loaded track: [`%s`](<%s>)", track.Info.Title, *track.Info.URI)),
			})
			tracksToQueue = []lavaqueue.QueueTrack{{Encoded: track.Encoded}}
		},
		func(playlist lavalink.Playlist) {
			_, _ = c.discordClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new(fmt.Sprintf("Loaded playlist: `%s` with `%d` tracks", playlist.Info.Name, len(playlist.Tracks))),
			})
			for _, t := range playlist.Tracks {
				tracksToQueue = append(tracksToQueue, lavaqueue.QueueTrack{Encoded: t.Encoded})
			}
		},
		func(tracks []lavalink.Track) {
			_, _ = c.discordClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new(fmt.Sprintf("Loaded search result: [`%s`](<%s>)", tracks[0].Info.Title, *tracks[0].Info.URI)),
			})
			tracksToQueue = []lavaqueue.QueueTrack{{Encoded: tracks[0].Encoded}}
		},
		func() {
			_, _ = c.discordClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new(fmt.Sprintf("Nothing found for: `%s`", identifier)),
			})
		},
		func(err error) {
			_, _ = c.discordClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new(fmt.Sprintf("Error while looking up query: `%s`", err)),
			})
		},
	))

	if len(tracksToQueue) == 0 {
		return nil
	}

	_, err := lavaqueue.AddQueueTracks(ctx, node, *e.GuildID(), tracksToQueue)
	return err
}
