package tempVoiceChannel

import (
	"context"
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
		fx.Provide(NewHandler, NewCleanupJob),
		fx.Provide(
			interactionCommand.AsCommand(NewInteractionCommand),
		),
		fx.Invoke(func(s *discordgo.Session, handler *Handler) {
			s.AddHandler(handler.Handle)
		}),
		fx.Invoke(func(scheduler gocron.Scheduler, job *CleanupJob, log *zap.Logger) error {
			_, err := scheduler.NewJob(
				gocron.DurationJob(5*time.Minute),
				gocron.NewTask(func() {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()

					if err := job.Run(ctx); err != nil {
						log.Error("failed to execute cleanup job", zap.Error(err))
					}
				}),
			)
			return err
		}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
