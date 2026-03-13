package musicPlayer

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	"github.com/SkinonikS/discord-bot-go/internal/v1/translator"
	disgobot "github.com/disgoorg/disgo/bot"
	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
	disgorest "github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgolink/v3/disgolink"
	disgolavalink "github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/lavaqueue-plugin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/fx"
)

const (
	MusicCommandName = "music"
)

type musicCommandImpl struct {
	t              translator.Translator
	lavaLinkClient disgolink.Client
	botClient      *disgobot.Client
	urlPattern     *regexp.Regexp
	searchPattern  *regexp.Regexp
}

type MusicCommandParams struct {
	fx.In

	T              translator.Translator
	LavaLinkClient disgolink.Client
	BotClient      *disgobot.Client
}

func NewMusicCommand(p MusicCommandParams) interactionCommand.Command {
	return &musicCommandImpl{
		t:              p.T,
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
		Name:                     c.Name(),
		NameLocalizations:        c.t.SimpleLocalizeAll(c.Name()),
		Description:              "Music player commands",
		DescriptionLocalizations: c.t.SimpleLocalizeAll("Music player commands"),
		Options: []disgodiscord.ApplicationCommandOption{
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:                     "queue",
				NameLocalizations:        c.t.SimpleLocalizeAll("queue"),
				Description:              "Show the current queue",
				DescriptionLocalizations: c.t.SimpleLocalizeAll("Show the current queue"),
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:                     "stop",
				NameLocalizations:        c.t.SimpleLocalizeAll("stop"),
				Description:              "Stop playback and disconnect",
				DescriptionLocalizations: c.t.SimpleLocalizeAll("Stop playback and disconnect"),
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:                     "skip",
				NameLocalizations:        c.t.SimpleLocalizeAll("skip"),
				Description:              "Skips the current song",
				DescriptionLocalizations: c.t.SimpleLocalizeAll("Skips the current song"),
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionInt{
						Name:                     "count",
						NameLocalizations:        c.t.SimpleLocalizeAll("count"),
						Description:              "The number of tracks to skip",
						DescriptionLocalizations: c.t.SimpleLocalizeAll("The number of tracks to skip"),
						Required:                 false,
					},
				},
			},
			disgodiscord.ApplicationCommandOptionSubCommand{
				Name:                     "play",
				NameLocalizations:        c.t.SimpleLocalizeAll("play"),
				Description:              "Play a track immediately",
				DescriptionLocalizations: c.t.SimpleLocalizeAll("Play a track immediately"),
				Options: []disgodiscord.ApplicationCommandOption{
					disgodiscord.ApplicationCommandOptionString{
						Name:                     "identifier",
						NameLocalizations:        c.t.SimpleLocalizeAll("identifier"),
						Description:              "Track search query or url",
						DescriptionLocalizations: c.t.SimpleLocalizeAll("Track search query or url"),
						Required:                 true,
					},
					disgodiscord.ApplicationCommandOptionString{
						Name:                     "source",
						NameLocalizations:        c.t.SimpleLocalizeAll("source"),
						Description:              "The source to search on",
						DescriptionLocalizations: c.t.SimpleLocalizeAll("The source to search on"),
						Required:                 false,
						Choices: []disgodiscord.ApplicationCommandOptionChoiceString{
							{
								Name:              "YouTube",
								NameLocalizations: c.t.SimpleLocalizeAll("YouTube"),
								Value:             string(disgolavalink.SearchTypeYouTube),
							},
							{
								Name:              "YouTube Music",
								NameLocalizations: c.t.SimpleLocalizeAll("YouTube Music"),
								Value:             string(disgolavalink.SearchTypeYouTubeMusic),
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
			Content: c.t.SimpleLocalize(e.Locale(), "No active player found."),
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
			Content: new(c.t.SimpleLocalize(e.Locale(), "Error while skipping track.")),
		}, disgorest.WithCtx(ctx))
		return err
	}

	if track == nil {
		_, err = e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
			Content: new(c.t.SimpleLocalize(e.Locale(), "No tracks in queue.")),
		}, disgorest.WithCtx(ctx))
		return err
	}

	_, err = e.Client().Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
		Content: new(c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
			MessageID: "Playing: {{.Title}}",
			TemplateData: map[string]any{
				"Title": track.Info.Title,
			},
		})),
	}, disgorest.WithCtx(ctx))
	return err
}

func (c *musicCommandImpl) handleStop(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error {
	player := c.lavaLinkClient.ExistingPlayer(*e.GuildID())
	if player == nil {
		return e.CreateMessage(disgodiscord.MessageCreate{
			Flags:   disgodiscord.MessageFlagEphemeral,
			Content: c.t.SimpleLocalize(e.Locale(), "No active player found."),
		}, disgorest.WithCtx(ctx))
	}

	if err := e.Client().UpdateVoiceState(ctx, *e.GuildID(), nil, false, false); err != nil {
		return fmt.Errorf("error while updating voice state: %w", err)
	}

	return e.CreateMessage(disgodiscord.MessageCreate{
		Flags:   disgodiscord.MessageFlagEphemeral,
		Content: c.t.SimpleLocalize(e.Locale(), "Playback stopped. Bot disconnected from voice channel."),
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
			Content: new(c.t.SimpleLocalize(e.Locale(), "No tracks in queue.")),
		}, disgorest.WithCtx(ctx))
		return err
	}

	var header string
	if queue.Type != "" {
		header = c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
			MessageID: "Queue ({{.Type}}):\n",
			TemplateData: map[string]any{
				"Type": queue.Type,
			},
		})
	} else {
		header = c.t.SimpleLocalize(e.Locale(), "Queue:\n")
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
		_, _ = tracksBuilder.WriteString(
			c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
				MessageID: "...and {{.Count}} more tracks.",
				TemplateData: map[string]any{
					"Count": remaining,
				},
			}),
		)
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
			Content: c.t.SimpleLocalize(e.Locale(), "You need to be in a voice channel to use this command."),
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
			content := c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
				MessageID: "Loaded track: [`{{.Title}}`](<{{.URI}}>)",
				TemplateData: map[string]any{
					"Title": track.Info.Title,
					"URI":   *track.Info.URI,
				},
			})

			_, _ = c.botClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new(content),
			})

			tracksToQueue = []lavaqueue.QueueTrack{{Encoded: track.Encoded, UserData: track.UserData}}
		},
		func(playlist disgolavalink.Playlist) {
			_, _ = c.botClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new(c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
					MessageID: "Loaded playlist: `{{.Name}}` with `{{.Count}}` tracks.",
					TemplateData: map[string]any{
						"Name":  playlist.Info.Name,
						"Count": len(playlist.Tracks),
					},
				})),
			})

			for i := range playlist.Tracks {
				t := &playlist.Tracks[i]
				tracksToQueue = append(tracksToQueue, lavaqueue.QueueTrack{Encoded: t.Encoded, UserData: t.UserData})
			}
		},
		func(tracks []disgolavalink.Track) {
			track := tracks[0]
			_, _ = c.botClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new(c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
					MessageID: "Loaded search result: [`{{.Title}}`](<{{.URI}}>).",
					TemplateData: map[string]any{
						"Title": track.Info.Title,
						"URI":   track.Info.URI,
					},
				})),
			})

			tracksToQueue = []lavaqueue.QueueTrack{{Encoded: track.Encoded, UserData: track.UserData}}
		},
		func() {
			_, _ = c.botClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new(c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
					MessageID: "Nothing found for: `{{.Query}}`.",
					TemplateData: map[string]any{
						"Query": identifier,
					},
				})),
			})
		},
		func(err error) {
			_, _ = c.botClient.Rest.UpdateInteractionResponse(e.ApplicationID(), e.Token(), disgodiscord.MessageUpdate{
				Content: new(c.t.Localize(e.Locale(), &i18n.LocalizeConfig{
					MessageID: "Error while looking up query: `{{.Error}}`.",
					TemplateData: map[string]any{
						"Error": err.Error(),
					},
				})),
			})
		},
	))

	if len(tracksToQueue) == 0 {
		return nil
	}

	_, err := lavaqueue.AddQueueTracks(ctx, node, *e.GuildID(), tracksToQueue)
	return err
}
