package migrator

import (
	"fmt"
	"sort"

	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
)

type RegistryParams struct {
	fx.In

	Providers []Provider `group:"migrator_providers"`
}

type Registry struct {
	providers map[string]*goose.Provider
}

func NewRegistry(p RegistryParams) (*Registry, error) {
	providers := make(map[string]*goose.Provider, len(p.Providers))
	for _, entry := range p.Providers {
		if _, exists := providers[entry.Name]; exists {
			return nil, fmt.Errorf("migrator: duplicate provider registered for store %q", entry.Name)
		}
		providers[entry.Name] = entry.Provider
	}

	return &Registry{providers: providers}, nil
}

func (r *Registry) Find(name string) (*goose.Provider, bool) {
	provider, ok := r.providers[name]
	return provider, ok
}

func (r *Registry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}
