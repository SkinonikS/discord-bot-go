package musicPlayerSource

import (
	"fmt"

	"go.uber.org/fx"
)

type Registry struct {
	sources map[string]Source
}

type RegistryParams struct {
	fx.In
	Sources []Source `group:"music_player_sources"`
}

func NewRegistry(p RegistryParams) (*Registry, error) {
	sources := make(map[string]Source)
	for _, src := range p.Sources {
		if _, ok := sources[src.Name()]; ok {
			return nil, fmt.Errorf("duplicate source name: %s", src.Name())
		}
		sources[src.Name()] = src
	}

	return &Registry{
		sources: sources,
	}, nil
}

func (r *Registry) Has(name string) bool {
	_, ok := r.sources[name]
	return ok
}

func (r *Registry) Find(name string) (Source, bool) {
	src, ok := r.sources[name]
	return src, ok
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.sources))
	for name := range r.sources {
		names = append(names, name)
	}
	return names
}
