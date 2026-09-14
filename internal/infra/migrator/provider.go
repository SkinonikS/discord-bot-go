package migrator

import (
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
)

type Provider struct {
	Name     string
	Provider *goose.Provider
}

func AsProvider(f any) any {
	return fx.Annotate(f, fx.ResultTags(`group:"migrator_providers"`))
}
