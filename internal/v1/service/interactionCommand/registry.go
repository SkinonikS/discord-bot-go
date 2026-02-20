package interactionCommand

import (
	"fmt"

	"github.com/SkinonikS/discord-bot-go/internal/v1/service/interactionCommand/command"
	"go.uber.org/fx"
)

type Registry struct {
	commands map[string]command.Command
}

type RegistryParams struct {
	fx.In
	Commands []command.Command `group:"discord_commands"`
}

func NewRegistry(p RegistryParams) (*Registry, error) {
	commands := make(map[string]command.Command)
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

func (r *Registry) Register(name string, command command.Command) error {
	if _, ok := r.commands[name]; ok {
		return fmt.Errorf("duplicate command name: %s", name)
	}

	r.commands[name] = command
	return nil
}

func (r *Registry) List() map[string]command.Command {
	return r.commands
}

func (r *Registry) Find(name string) (command.Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}
