package interactioncommand

import (
	"fmt"
	"sync"

	"github.com/samber/lo"
	"go.uber.org/fx"
)

type Registry interface {
	List() []Command
	ListByScope(scope CommandScope) []Command
	Find(name string) (Command, bool)
	Has(name string) bool
	Register(cmd Command) error
}

type registryImpl struct {
	mu       sync.RWMutex
	commands map[string]Command
}

func NewRegistry() Registry {
	return &registryImpl{
		commands: make(map[string]Command),
	}
}

type RegistryParams struct {
	fx.In

	Commands []Command `group:"discord_commands"`
}

func populateRegistry(p RegistryParams, registry Registry) error {
	for _, cmd := range p.Commands {
		if err := registry.Register(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (r *registryImpl) List() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return lo.Values(r.commands)
}

func (r *registryImpl) ListByScope(scope CommandScope) []Command {
	return lo.Filter(r.List(), func(cmd Command, _ int) bool {
		return cmd.Scope() == scope
	})
}

func (r *registryImpl) Find(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmd, ok := r.commands[name]
	return cmd, ok
}

func (r *registryImpl) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.commands[name]
	return ok
}

func (r *registryImpl) Register(cmd Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.commands[cmd.Name()]; ok {
		return fmt.Errorf("duplicate command name: %s", cmd.Name())
	}

	r.commands[cmd.Name()] = cmd
	return nil
}
