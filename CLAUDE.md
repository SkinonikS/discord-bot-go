# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

Build via `just` (not `make` — there is no Makefile):

```bash
just build-bot            # build ./bot binary (cmd/bot/main.go)
just build-cli             # build ./cli binary (cmd/cli/main.go)
just create-migration name # generate a new goose SQL migration under migrations/
```

Run directly:

```bash
CGO_ENABLED=1 go run ./cmd/bot/main.go
CGO_ENABLED=1 go run ./cmd/cli/main.go migrate rollback
```

Test and lint (CI in `.github/workflows/test.yml` runs both against live Postgres + Lavalink service containers):

```bash
go test -v -race -count=1 ./...   # run all tests
go test -race -run TestName ./internal/service/auto_role/...  # single test
golangci-lint run                  # lint (config: .golangci.yml)
```

Tests that hit the database/Lavalink expect these env vars (see `.env.example` / `docker-compose.yml` for local defaults):
`DB_POSTGRES_HOST`, `DB_POSTGRES_PORT`, `DB_POSTGRES_USER`, `DB_POSTGRES_PASSWORD`, `DB_POSTGRES_DB`, `DB_POSTGRES_SSLMODE`, `LAVALINK_MAIN_NODE_HOST`, `LAVALINK_MAIN_NODE_PORT`, `LAVALINK_MAIN_NODE_PASSWORD`, `LAVALINK_MAIN_NODE_SECURE`.

CGO is required (the bot links `libopus` and `libdave` for Discord's DAVE voice encryption via `github.com/disgoorg/godave`). See README.md's "CGO Dependencies" section for install steps; Docker builds handle this automatically.

Database migrations run automatically on bot startup (goose, via `internal/infra/database/migrator`); the `cli` binary is only needed for rollbacks or ad-hoc migration commands.

## Architecture

This is a Discord bot built entirely around **Uber Fx** dependency injection. There is no code that isn't wired through an `fx.Module` — when adding anything new, the pattern below is not optional style, it's how the app boots.

### Layering: `internal/infra` vs `internal/service`

- **`internal/infra/*`** — infrastructure/platform modules (config, logger, database, discord client, lavalink, cron, cache, translator, http_server, readiness, foundation). Each is an isolated Go package with its own `fx.go` exposing `NewModule() fx.Option`, plus a `config.go` reading its own YAML section via `go.uber.org/config`.
- **`internal/service/*`** — feature/business modules (`auto_role`, `reaction_role`, `temp_voice_channel`, `music_player`, `interaction_command`). Each depends on one or more infra modules but never on another service module directly except `interaction_command` (which all command-registering services depend on to register their slash commands).

Both app entrypoints assemble modules the same way — see `internal/app/bot/app.go` (the bot: infra modules, then `service.NewModule()`) and `internal/app/cli/internal/fx.go` (the CLI: a narrower slice of infra needed for migrations). `internal/service/fx.go` is the single place that lists which feature modules are active.

### The per-module fx.go pattern

Every module (infra or service) follows this shape — copy it when adding a new one:

```go
const ModuleName = "..."

func NewModule() fx.Option {
    return fx.Module(ModuleName,
        fx.Provide(NewConfig, NewService, NewRepo, ...),
        fx.Provide(
            discord.AsEventListener(NewEventListener),
            interaction_command.AsCommand(NewFooCommand),
        ),
        fx.Decorate(func(log *zap.Logger) *zap.Logger {
            return log.Named(ModuleName)
        }),
    )
}
```

The `AsX(f any) any` helper pattern (`discord.AsEventListener`, `interaction_command.AsCommand`, `readiness.AsHandler`) wraps a constructor with `fx.Annotate(..., fx.As(new(Interface)), fx.ResultTags(`group:"..."`))` so it self-registers into a value group. Consumers collect the group with an `fx.In`-embedded params struct field tagged `group:"..."` (see `interaction_command.RegistryParams`, `readiness.HandlerParams`). This is how event listeners, slash commands, and readiness checks are discovered without any central registration list — to add a new slash command or event listener, just provide it with the right `AsX` wrapper from your module's `fx.go`, nothing else needs to change.

### Config

Config is layered YAML + env expansion, not plain env vars:

- `config/config.yaml` (or `config/config.<env>.yaml`) has `${VAR:default}` placeholders expanded against the process environment (`internal/infra/config/config.go`, using `go.uber.org/config`).
- `.env` (or `.env.<env>`) is loaded first via `godotenv` to populate those env vars — see `.env.example` for the full variable list.
- Each infra module reads its own top-level YAML key into its own `*Config` struct via `provider.Get(ConfigKey).Populate(...)` (e.g. `internal/infra/database/config.go` reads the `database:` key). Follow this pattern for any new configurable module rather than reading `config.yaml` directly.

### Feature module shape (e.g. `internal/service/auto_role`)

A typical feature module has: `service.go` (business logic interface + impl, depends on a `Repo` and possibly the Discord REST client), `repo.go` (GORM-backed persistence), `eventListener.go` (implements `disgobot.EventListener`, registered via `discord.AsEventListener`), and a `*Command.go` (implements `interaction_command.Command`, registered via `interaction_command.AsCommand`). Slash commands are collected into a `Registry` and registered globally with Discord on bot startup (see `interaction_command/fx.go`'s `fx.Invoke` hook).

### Persistence

GORM over Postgres (`internal/infra/database`). Migrations are plain SQL files under `migrations/`, managed by `goose` (`github.com/pressly/goose/v3`) — create new ones with `just create-migration <name>`, they run automatically when the bot starts.

### i18n

`internal/infra/translator` wraps `go-i18n`; per-locale JSON message files live in `i18n/`. Slash command responses are localized via the `Translator` interface rather than hardcoded strings.

### Deployment

`.github/workflows/test.yml` runs on every push/PR (lint + test against live Postgres/Lavalink containers). `.github/workflows/deploy.yml` additionally builds and pushes the bot Docker image and deploys to Kubernetes via Kustomize (`k8s/base` + `k8s/overlays/prod`) on pushes to `main` (skipped for tag pushes, per commit `cbcbdd9`).
