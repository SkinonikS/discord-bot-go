package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/SkinonikS/discord-bot-go/internal/v1/foundation"
	disgodiscord "github.com/disgoorg/disgo/discord"
	"github.com/gookit/goutil/fsutil"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

type Translator interface {
	LocalizeAll(lc *i18n.LocalizeConfig) map[disgodiscord.Locale]string
	SimpleLocalizeAll(msg string) map[disgodiscord.Locale]string
	SimpleLocalize(locale disgodiscord.Locale, msg string) string
	LocalizeMessage(locale disgodiscord.Locale, msg *i18n.Message) string
	Localize(locale disgodiscord.Locale, lc *i18n.LocalizeConfig) string
	LoadTranslations(locales ...disgodiscord.Locale) error
}

type translatorImpl struct {
	log           *zap.SugaredLogger
	path          *foundation.Path
	bundle        *i18n.Bundle
	localizers    map[disgodiscord.Locale]*i18n.Localizer
	defaultLocale disgodiscord.Locale
}

type Params struct {
	fx.In

	Config *Config
	Path   *foundation.Path
	Log    *zap.Logger
}

func New(p Params) Translator {
	defaultLocale := p.Config.DefaultLocale
	if defaultLocale.String() == "unknown" {
		defaultLocale = disgodiscord.LocaleEnglishUS
	}

	return &translatorImpl{
		defaultLocale: defaultLocale,
		log:           p.Log.Sugar(),
		path:          p.Path,
		bundle:        i18n.NewBundle(language.Make(string(defaultLocale))),
		localizers:    make(map[disgodiscord.Locale]*i18n.Localizer),
	}
}

func (t *translatorImpl) SimpleLocalizeAll(msg string) map[disgodiscord.Locale]string {
	result := make(map[disgodiscord.Locale]string, len(t.bundle.LanguageTags()))
	for _, tag := range t.bundle.LanguageTags() {
		locale := disgodiscord.Locale(tag.String())
		result[locale] = t.SimpleLocalize(locale, msg)
	}

	return result
}

func (t *translatorImpl) LocalizeAll(lc *i18n.LocalizeConfig) map[disgodiscord.Locale]string {
	result := make(map[disgodiscord.Locale]string, len(t.bundle.LanguageTags()))
	for _, tag := range t.bundle.LanguageTags() {
		locale := disgodiscord.Locale(tag.String())
		result[locale] = t.Localize(locale, lc)
	}

	return result
}

func (t *translatorImpl) SimpleLocalize(locale disgodiscord.Locale, msg string) string {
	return t.Localize(locale, &i18n.LocalizeConfig{
		MessageID: msg,
	})
}

func (t *translatorImpl) LocalizeMessage(locale disgodiscord.Locale, msg *i18n.Message) string {
	return t.Localize(locale, &i18n.LocalizeConfig{
		DefaultMessage: msg,
	})
}

func (t *translatorImpl) Localize(locale disgodiscord.Locale, lc *i18n.LocalizeConfig) string {
	result, err := t.makeLocalizer(locale).Localize(lc)
	if err != nil {
		var messageID string
		if lc.MessageID != "" {
			messageID = lc.MessageID
		} else if lc.DefaultMessage != nil {
			messageID = lc.DefaultMessage.ID
		}

		return t.fallbackRenderTemplate(messageID, lc.TemplateData)
	}

	return result
}

func (t *translatorImpl) LoadTranslations(locales ...disgodiscord.Locale) error {
	var failed []string
	for _, locale := range lo.Uniq(locales) {
		fileName := string(locale) + ".json"
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

	for _, locale := range lo.Uniq(locales) {
		_ = t.bundle.AddMessages(language.Make(string(locale)))
	}

	if len(failed) > 0 {
		return fmt.Errorf("failed to load translations for locales: %s", strings.Join(failed, ", "))
	}
	return nil
}

func (t *translatorImpl) makeLocalizer(locale disgodiscord.Locale) *i18n.Localizer {
	loc, ok := t.localizers[locale]
	if !ok {
		loc = i18n.NewLocalizer(t.bundle, string(locale))
		t.localizers[locale] = loc
	}

	return loc
}

func (*translatorImpl) loadTranslations(buf []byte, path string) (*i18n.MessageFile, error) {
	file, err := i18n.ParseMessageFileBytes(buf, path, map[string]i18n.UnmarshalFunc{
		"json": json.Unmarshal,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse translation file: %w", err)
	}

	return file, nil
}

func (*translatorImpl) fallbackRenderTemplate(tmpl string, data any) string {
	if data == nil {
		return tmpl
	}

	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return tmpl
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return tmpl
	}

	return buf.String()
}
