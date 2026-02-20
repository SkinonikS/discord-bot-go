package tempVoiceChannel

import (
	"time"

	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand"
	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "tempVoiceChannel"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewHandler, NewJob, NewInteractionCommand),
		fx.Invoke(func(s gocron.Scheduler, job *Job) error {
			_, err := s.NewJob(
				gocron.DurationJob(1*time.Second),
				gocron.NewTask(job.Run),
			)
			return err
		}),
		fx.Invoke(func(c *interactionCommand.Registry, cmd *InteractionCommand) error {
			return c.Register(cmd.Name(), cmd)
		}),
		fx.Invoke(func(s *discordgo.Session, handler *Handler) {
			s.AddHandler(handler.Handle)
		}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
