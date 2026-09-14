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
}

type registryImpl struct {
	mu       sync.RWMutex
	commands map[string]Command
}

type RegistryParams struct {
	fx.In

	Commands []Command `group:"discord_commands"`
}

func NewRegistry() Registry {
	return &registryImpl{
		commands: make(map[string]Command),
	}
}

func populateRegistry(registry Registry, p RegistryParams) error {
	r, ok := registry.(*registryImpl)
	if !ok {
		return fmt.Errorf("unexpected registry implementation: %T", registry)
	}

	commands := make(map[string]Command, len(p.Commands))
	for _, cmd := range p.Commands {
		if _, ok := commands[cmd.Name()]; ok {
			return fmt.Errorf("duplicate command name: %s", cmd.Name())
		}

		commands[cmd.Name()] = cmd
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands = commands

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
