package interactioncommand

import (
	"context"

	disgodiscord "github.com/disgoorg/disgo/discord"
	disgoevents "github.com/disgoorg/disgo/events"
)

type CommandScope int

const (
	// CommandScopeGuild commands are registered per-guild, so each guild can independently
	// enable or disable them.
	CommandScopeGuild CommandScope = iota
	// CommandScopeGlobal commands are registered once for every guild the bot is in and
	// cannot be disabled per-guild. Reserved for system commands (e.g. ping, info).
	CommandScopeGlobal
)

type Command interface {
	Definition() disgodiscord.SlashCommandCreate
	Execute(ctx context.Context, e *disgoevents.ApplicationCommandInteractionCreate) error
	Name() string
	Scope() CommandScope
}
