package interactionCommand

import (
	"fmt"

	"github.com/samber/lo"
	"go.uber.org/fx"
)

type Registry struct {
	commands map[string]Command
}

type RegistryParams struct {
	fx.In
	Commands []Command `group:"discord_commands"`
}

func NewRegistry(p RegistryParams) (*Registry, error) {
	commands := make(map[string]Command)
	for _, cmd := range p.Commands {
		if _, ok := commands[cmd.Name()]; ok {
			return nil, fmt.Errorf("duplicate command name: %s", cmd.Name())
		}

		commands[cmd.Name()] = cmd
	}

	return &Registry{
		commands: commands,
	}, nil
}

func (r *Registry) Register(cmd Command) error {
	if _, ok := r.commands[cmd.Name()]; ok {
		return fmt.Errorf("duplicate command name: %s", cmd.Name())
	}
	r.commands[cmd.Name()] = cmd
	return nil
}

func (r *Registry) List() []Command {
	return lo.Values(r.commands)
}

func (r *Registry) Find(name string) (Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}
