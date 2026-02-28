package interactionCommand

import (
	"fmt"
	"slices"

	"github.com/SkinonikS/discord-bot-go/internal/v1/discord"
	"github.com/SkinonikS/discord-bot-go/internal/v1/util"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Handler struct {
	log      *zap.SugaredLogger
	commands *Registry
}

type HandlerParams struct {
	fx.In
	Log      *zap.Logger
	Commands *Registry
}

func NewHandler(p HandlerParams) *Handler {
	return &Handler{
		log:      p.Log.Sugar(),
		commands: p.Commands,
	}
}

func (h *Handler) Handle(s *discordgo.Session, e *discordgo.InteractionCreate) {
	err := util.Safe(func() error {
		ctx, cancel := discord.DefaultHandlerContext()
		defer cancel()

		if !h.isApplicable(e.Interaction) {
			return nil
		}

		cmdData := e.ApplicationCommandData()
		cmd, ok := h.commands.Find(cmdData.Name)
		if !ok {
			h.log.Warnw("unknown command executed", zap.String("command", cmdData.Name))
			return nil
		}

		if err := cmd.Execute(ctx, s, e); err != nil {
			return fmt.Errorf("failed to handle command: %w", err)
		}

		return nil
	})
	if err != nil {
		h.notifyUserAboutError(s, e, err)
		h.log.Errorw("failed to handle interaction", zap.Error(err))
	}
}

func (h *Handler) isApplicable(i *discordgo.Interaction) bool {
	return slices.Contains([]discordgo.InteractionType{
		discordgo.InteractionApplicationCommand,
		discordgo.InteractionApplicationCommandAutocomplete,
	}, i.Type)
}

func (h *Handler) notifyUserAboutError(s *discordgo.Session, e *discordgo.InteractionCreate, err error) {
	embed := &discordgo.MessageEmbed{
		Title:       "Execution Failed",
		Description: "Something went wrong while executing this command.\n```" + err.Error() + "```",
		Color:       0xff0000,
	}

	if err := s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		h.log.Errorw("unable to notify user about error", zap.Error(err))
	}
}
