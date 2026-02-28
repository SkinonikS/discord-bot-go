package translator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/goutil/fsutil"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

type Translator struct {
	log        *zap.SugaredLogger
	path       *foundation.Path
	bundle     *i18n.Bundle
	localizers map[discordgo.Locale]*i18n.Localizer
}

type Params struct {
	fx.In
	Config *Config
	Path   *foundation.Path
	Log    *zap.Logger
}

func New(p Params) *Translator {
	return &Translator{
		log:        p.Log.Sugar(),
		bundle:     i18n.NewBundle(language.Make(p.Config.DefaultLocale.String())),
		path:       p.Path,
		localizers: make(map[discordgo.Locale]*i18n.Localizer),
	}
}

func (t *Translator) LocalizeAll(lc i18n.LocalizeConfig) map[discordgo.Locale]string {
	result := make(map[discordgo.Locale]string, len(t.bundle.LanguageTags()))
	for _, tag := range t.bundle.LanguageTags() {
		locale := discordgo.Locale(tag.String())
		result[locale] = t.Localize(locale, lc)
	}
	return result
}

func (t *Translator) SimpleLocalize(locale discordgo.Locale, msg string) string {
	return t.Localize(locale, i18n.LocalizeConfig{
		MessageID: msg,
	})
}

func (t *Translator) LocalizeMessage(locale discordgo.Locale, msg i18n.Message) string {
	return t.Localize(locale, i18n.LocalizeConfig{
		DefaultMessage: &msg,
	})
}

func (t *Translator) Localize(locale discordgo.Locale, lc i18n.LocalizeConfig) string {
	result, err := t.makeLocalizer(locale).Localize(&lc)
	if err != nil {
		if lc.DefaultMessage != nil {
			return lc.DefaultMessage.ID
		}
		return lc.MessageID
	}
	return result
}

func (t *Translator) LoadTranslations(locales ...discordgo.Locale) error {
	var failed []string
	for _, locale := range lo.Uniq(locales) {
		fileName := fmt.Sprintf("%s.json", string(locale))
		fullPath := t.path.I18nPath(fileName)

		if !fsutil.IsFile(fullPath) {
			failed = append(failed, string(locale))
			continue
		}

		buf, err := os.ReadFile(fullPath)
		if err != nil {
			failed = append(failed, string(locale))
			continue
		}

		messages, err := t.loadTranslations(buf, fullPath)
		if err != nil {
			t.log.Warnf("failed to load translation file: %s", err)
			failed = append(failed, string(locale))
			continue
		}

		if err := t.bundle.AddMessages(language.Make(string(locale)), messages.Messages...); err != nil {
			t.log.Warnf("failed to add translation messages: %s", err)
			failed = append(failed, string(locale))
			continue
		}

		if _, err := t.bundle.LoadMessageFile(fullPath); err != nil {
			t.log.Warnf("failed to load translation file: %s", err)
			failed = append(failed, string(locale))
			continue
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("failed to load translations for locales: %s", strings.Join(failed, ", "))
	}
	return nil
}

func (t *Translator) makeLocalizer(locale discordgo.Locale) *i18n.Localizer {
	loc, ok := t.localizers[locale]
	if !ok {
		loc = i18n.NewLocalizer(t.bundle, string(locale))
		t.localizers[locale] = loc
	}
	return loc
}

func (*Translator) loadTranslations(buf []byte, path string) (*i18n.MessageFile, error) {
	return i18n.ParseMessageFileBytes(buf, path, map[string]i18n.UnmarshalFunc{
		"json": json.Unmarshal,
	})
}
