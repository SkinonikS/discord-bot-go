package musicPlayer

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	disgobot "github.com/disgoorg/disgo/bot"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgolink/v3/disgolink"
	disgolavalink "github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/lavaqueue-plugin"
	"go.uber.org/fx"
)

const (
	MusicCommandName = "music"
)

type musicCommandImpl struct {
	lavaLinkClient disgolink.Client
	botClient      *disgobot.Client
	urlPattern     *regexp.Regexp
	searchPattern  *regexp.Regexp
}

type MusicCommandParams struct {
	fx.In

	LavaLinkClient disgolink.Client
	BotClient      *disgobot.Client
}

func NewMusicCommand(p MusicCommandParams) interactionCommand.Command {
	return &musicCommandImpl{
		lavaLinkClient: p.LavaLinkClient,
		botClient:      p.BotClient,
		urlPattern:     regexp.MustCompile("^https?://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]?"),
		searchPattern:  regexp.MustCompile(`^(.{2})search:(.+)`),
	}
}

func (c *musicCommandImpl) Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
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

func (c *musicCommandImpl) Definition() disgodiscord.SlashCommandCreate {
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
								Value: string(disgolavalink.SearchTypeYouTube),
							},
							{
								Name:  "YouTube Music",
								Value: string(disgolavalink.SearchTypeYouTubeMusic),
							},
						},
					},
				},
			},
		},
	}
}

func (c *musicCommandImpl) Name() string {
	return MusicCommandName
}

func (c *musicCommandImpl) handleSkip(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
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

	count := max(data.Int("count"), 1)

	track, err := lavaqueue.QueueNextTrack(ctx, player.Node(), *e.GuildID(), count)
	if err != nil {
		_, err = e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
			Content: new("error while skipping track"),
		}, disgorest.WithCtx(ctx))
		return err
	}

	if track == nil {
		_, err = e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
			Content: new("no tracks in queue"),
		}, disgorest.WithCtx(ctx))
		return err
	}

	_, err = e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
		Content: new("Playing: " + track.Info.Title),
	}, disgorest.WithCtx(ctx))
	return err
}

func (c *musicCommandImpl) handleStop(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
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
		Content: "Playback stopped. Bot disconnected from voice channel.",
	})
}

func (c *musicCommandImpl) handleQueue(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	queue, err := lavaqueue.GetQueue(ctx, c.lavaLinkClient.BestNode(), *e.GuildID())
	if err != nil {
		return err
	}

	if len(queue.Tracks) == 0 {
		_, err = c.botClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
			Content: new("No tracks in queue."),
		}, disgorest.WithCtx(ctx))
		return err
	}

	header := "Queue:\n"
	if queue.Type != "" {
		header = fmt.Sprintf("Queue `%s`:\n", queue.Type)
	}

	const maxLen = 1900
	tracksBuilder := &strings.Builder{}
	tracksBuilder.WriteString(header)
	shown := 0
	for i := range queue.Tracks {
		t := &queue.Tracks[i]
		line := fmt.Sprintf("%d. [`%s`](<%s>)\n", i+1, t.Info.Title, *t.Info.URI)
		if tracksBuilder.Len()+len(line) > maxLen {
			break
		}

		tracksBuilder.WriteString(line)
		shown++
	}

	if remaining := len(queue.Tracks) - shown; remaining > 0 {
		_, _ = fmt.Fprintf(tracksBuilder, "...and %d more tracks", remaining)
	}

	_, err = c.botClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
		Content: new(tracksBuilder.String()),
	}, disgorest.WithCtx(ctx))
	return err
}

func (c *musicCommandImpl) handlePlay(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()

	identifier := data.String("identifier")
	if source, ok := data.OptString("source"); ok {
		identifier = disgolavalink.SearchType(source).Apply(identifier)
	} else if !c.urlPattern.MatchString(identifier) && !c.searchPattern.MatchString(identifier) {
		identifier = disgolavalink.SearchTypeYouTube.Apply(identifier)
	}

	voiceState, ok := c.botClient.Caches.VoiceState(*e.GuildID(), e.User().ID)
	if !ok {
		return e.CreateMessage(disgodiscord.MessageCreate{
			Flags:   disgodiscord.MessageFlagEphemeral,
			Content: "You need to be in a voice channel to use this command.",
		})
	}

	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	if err := c.botClient.UpdateVoiceState(ctx, *e.GuildID(), voiceState.ChannelID, false, true); err != nil {
		return err
	}

	node := c.lavaLinkClient.BestNode()

	var tracksToQueue []lavaqueue.QueueTrack
	node.LoadTracksHandler(ctx, identifier, disgolink.NewResultHandler(
		func(track disgolavalink.Track) {
			_, _ = c.botClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new(fmt.Sprintf("Loaded track: [`%s`](<%s>)", track.Info.Title, *track.Info.URI)),
			})

			tracksToQueue = []lavaqueue.QueueTrack{{Encoded: track.Encoded, UserData: track.UserData}}
		},
		func(playlist disgolavalink.Playlist) {
			_, _ = c.botClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new(fmt.Sprintf("Loaded playlist: `%s` with `%d` tracks", playlist.Info.Name, len(playlist.Tracks))),
			})

			for i := range playlist.Tracks {
				t := &playlist.Tracks[i]
				tracksToQueue = append(tracksToQueue, lavaqueue.QueueTrack{Encoded: t.Encoded, UserData: t.UserData})
			}
		},
		func(tracks []disgolavalink.Track) {
			track := tracks[0]

			_, _ = c.botClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new(fmt.Sprintf("Loaded search result: [`%s`](<%s>)", track.Info.Title, *track.Info.URI)),
			})

			tracksToQueue = []lavaqueue.QueueTrack{{Encoded: track.Encoded, UserData: track.UserData}}
		},
		func() {
			_, _ = c.botClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new("Nothing found for: `" + identifier + "`"),
			})
		},
		func(err error) {
			_, _ = c.botClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new("Error while looking up query: `" + err.Error() + "`"),
			})
		},
	))

	if len(tracksToQueue) == 0 {
		return nil
	}

	_, err := lavaqueue.AddQueueTracks(ctx, node, *e.GuildID(), tracksToQueue)
	return err
}
