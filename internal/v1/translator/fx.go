package translator

import (
	"github.com/bwmarrin/discordgo"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	ModuleName = "translator"
)

func NewModule() fx.Option {
	return fx.Module(ModuleName,
		fx.Provide(NewConfig, New),
		fx.Invoke(func(t *Translator, cfg *Config, log *zap.Logger) {
			locales := make(map[discordgo.Locale]struct{})
			if cfg.DefaultLocale.String() == "unknown" {
				log.Warn("default locale is unknown, fallback language will be used", zap.String("locale", string(cfg.DefaultLocale)))
				cfg.DefaultLocale = discordgo.EnglishUS
			}
			locales[cfg.DefaultLocale] = struct{}{}

			for _, locale := range cfg.AvailableLocales {
				if locale.String() == "unknown" {
					log.Warn("unsupported discord locale", zap.String("locale", string(locale)))
					continue
				}
				locales[locale] = struct{}{}
			}

			if err := t.LoadTranslations(lo.Keys(locales)...); err != nil {
				log.Warn("failed to load translation file", zap.Error(err))
				return
			}

			log.Info("translations loaded", zap.String("defaultLocale", cfg.DefaultLocale.String()))
		}),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named(ModuleName)
		}),
	)
}
