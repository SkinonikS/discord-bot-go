# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

Build via `just` (not `make` — there is no Makefile):

```bash
just build-bot               # build ./bot binary (cmd/bot/main.go)
just build-cli                # build ./cli binary (cmd/cli/main.go)
just create-migration name    # generate a new goose SQL migration under migrations/postgres/
```

Run directly:

```bash
CGO_ENABLED=1 go run ./cmd/bot/main.go
CGO_ENABLED=1 go run ./cmd/cli/main.go migrate up --store postgres        # apply pending migrations for a store
CGO_ENABLED=1 go run ./cmd/cli/main.go migrate rollback --store postgres  # roll back the last migration for a store
CGO_ENABLED=1 go run ./cmd/cli/main.go commands sync                      # register global + guild commands with Discord
CGO_ENABLED=1 go run ./cmd/cli/main.go commands list --guild <id>         # list guild-scoped commands and enabled/disabled status for a guild
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

Setting `APP_DEBUG=true` enables a `pprof` server (`gin-contrib/pprof`) on `:6060` by default (override with `PPROF_ADDR`) for local profiling.

The bot does **not** apply migrations or register commands itself. Both are applied with the `cli` binary, one store at a time for migrations — `cli migrate up --store <name>` applies pending migrations, `cli migrate rollback --store <name>` rolls one back (both default `--store` to `postgres`, currently the only registered store) — and account-wide for commands via `cli commands sync`. In the deploy pipeline these run as a dedicated Kubernetes Job (see "Deployment" below) before the bot rolls out; locally, run them yourself after `just create-migration` or before starting the bot for the first time.

### Code generation

Repository query code and test mocks are generated, not hand-written — after changing a `query.sql` file or a migration's schema, or an interface that's mocked, regenerate rather than editing generated output by hand:

```bash
sqlc generate   # regenerates internal/service/repository/*/postgres/internal/gen from query.sql + migrations/postgres (config: sqlc.yaml)
mockery         # regenerates */mock packages from the interfaces listed in .mockery.yml
```

## Architecture

This is a Discord bot built entirely around **Uber Fx** dependency injection. There is no code that isn't wired through an `fx.Module` — when adding anything new, the pattern below is not optional style, it's how the app boots.

### Layering: `internal/infra` vs `internal/service`

- **`internal/infra/*`** — infrastructure/platform modules (`config`, `logger`, `postgres`, `redis`, `discord`, `lavalink`, `migrator`, `cron`, `cache`, `translator`, `http_server`, `readiness`, `foundation`). Each is an isolated Go package with its own `fx.go` exposing `NewModule() fx.Option`, plus a `config.go` reading its own YAML section via `go.uber.org/config`.
- **`internal/service/*`** — feature/business modules (`auto_role`, `reaction_role`, `temp_voice_channel`, `music_player`, `interaction_command`). Each depends on one or more infra modules but never on another service module directly except `interaction_command` (which all command-registering services depend on to register their slash commands).

There are two binaries, each assembling its own `fx.New(...)` graph from these modules — see `internal/app/bot/app.go` (the full bot: gateway + HTTP/readiness server + Lavalink + Redis + `service.NewModule()`) and `internal/app/cli/app.go` (migrations + command sync/list — a narrower slice of infra, but it still wires the full `service.NewModule()` because `cli commands sync`/`list` need the same `Registry`/`Settings`/`Syncer` the bot builds; see the value-group gotcha below for why that drags in things like `lavalink`). `internal/service/fx.go` is the single place that lists which feature modules are active (it also wires `repository.NewModule()`, which centralizes binding every Postgres repo implementation to its interface via `fx.Annotate(..., fx.As(...))`).

The `cli` binary's command tree is itself one more layer of fx modules on top of the above: `internal/app/cli/app.go` wires `internal/app/cli/internal` (`appservice.NewModule()`), which wires `internal/app/cli/internal/cli` (`cli.NewModule()`, module name `"urfave"`) — this is where the `urfave/cli` commands (`migrate`, `commands`) are built and collected via a `cli_commands` value group, then handed to the root `*cli.Command` with `fx.Populate(&cmd)`. New `cli` subcommands follow the same `AsX` value-group pattern as everything else.

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
            cron.AsJob(NewFooCronJob),
        ),
        fx.Decorate(func(log *zap.Logger) *zap.Logger {
            return log.Named(ModuleName)
        }),
    )
}
```

The `AsX(f any) any` helper pattern (`discord.AsEventListener`, `interaction_command.AsCommand`, `cron.AsJob`, `readiness.AsHandler`) wraps a constructor with `fx.Annotate(..., fx.As(new(Interface)), fx.ResultTags(`group:"..."`))` so it self-registers into a value group. Consumers collect the group with an `fx.In`-embedded params struct field tagged `group:"..."` (see `interaction_command.RegistryParams`, `readiness.HandlerParams`). This is how event listeners, slash commands, and readiness checks are discovered without any central registration list — to add a new one, just provide it with the right `AsX` wrapper from your module's `fx.go`, nothing else needs to change.

**Gotcha:** the consumers of a value group are wired via `fx.Invoke` (e.g. `discord.NewModule()`'s worker pool, `interaction_command.NewModule()`'s `populateRegistry`), so they build *every* provider in the group eagerly on startup, regardless of which app is running. This means any binary that wires `service.NewModule()` transitively needs every dependency that every registered command/listener needs to construct — e.g. `music_player`'s slash command needs `disgolink.Client`, so `lavalink.NewModule()` must be wired even into the `cli` binary, which never plays audio or opens a gateway connection. Forgetting this surfaces as an fx "missing type" error at the group's `fx.Invoke` site, not at the module that actually needs the dependency.

**Gotcha (current state):** `cron.NewModule()` — which provides the `gocron.Scheduler` and, via its own `fx.Invoke`, is what actually forces construction of every `cron.AsJob`-registered job — is not wired into either app graph (`bot` or `cli`) right now. `reaction_role`'s stale-entries job and `temp_voice_channel`'s orphaned-channel job are registered into the `cron_jobs` group but nothing collects that group, so neither job is currently constructed or scheduled to run. A previous standalone `scheduler` binary/app (which did wire `cron.NewModule()`) has been removed from the tree; whatever replaces it needs to wire `cron.NewModule()` somewhere.

### Config

Config is layered YAML + env expansion, not plain env vars:

- `config/config.yaml` (or `config/config.<env>.yaml`) has `${VAR:default}` placeholders expanded against the process environment (`internal/infra/config/config.go`, using `go.uber.org/config`).
- `.env` (or `.env.<env>`) is loaded first via `godotenv` to populate those env vars — see `.env.example` for the full variable list.
- Each infra module reads its own top-level YAML key into its own `*Config` struct via `provider.Get(ConfigKey).Populate(...)` (e.g. `internal/infra/postgres/config.go` reads the `postgres:` key). Follow this pattern for any new configurable module rather than reading `config.yaml` directly.

### Feature module shape (e.g. `internal/service/auto_role`)

A typical feature module has: `service.go` (business logic interface + impl, depends on a `Repo` and possibly the Discord REST client — obtained as `disgobot.Client.Rest`, part of `internal/infra/discord`'s `Result`, not a separate module), `event_listener.go` (implements `disgobot.EventListener`, registered via `discord.AsEventListener`), and a `*_command.go` (implements `interaction_command.Command`, registered via `interaction_command.AsCommand`). `reaction_role` and `temp_voice_channel` additionally have a `*_job.go` implementing `cron.Job` (a `Definition() gocron.JobDefinition` + `Task() gocron.Task` pair, registered via `cron.AsJob` — see the cron gotcha above for why these don't currently run).

### Slash commands: registry, per-guild settings, and sync

`internal/service/interaction_command` has three collaborating pieces beyond the `Registry` (which just collects every `Command` from the `discord_commands` group):

- **`Command.Scope()`** is either `CommandScopeGlobal` (system commands like `ping`/`info`/the `commands` management command itself — always registered, can't be disabled per guild) or `CommandScopeGuild` (feature commands — can be toggled off per guild).
- **`Settings`** (`settings.go`) wraps the `disabled_guild_commands` repo (Postgres-backed, `disabled_guild_commands` table) to record which `CommandScopeGuild` commands are disabled for a given guild.
- **`Syncer`** (`syncer.go`) pushes definitions to Discord over REST: `SyncGlobalCommands` (all `CommandScopeGlobal` commands, account-wide), `SyncGuildCommands(guildID)` (that guild's `CommandScopeGuild` commands, filtered by its `Settings`), and `SyncAllGuildCommands` (iterates every guild the bot is currently in).

The `commands` slash command (`manage_command.go`, admin-only, guild context) exposes `list`/`enable`/`disable` subcommands backed by `Settings` + `Syncer`, so a server admin can turn a feature off without redeploying. The bot's event listener also calls `Syncer.SyncGuildCommands` on `GuildJoin`. The `cli commands sync` / `cli commands list --guild <id>` subcommands (`internal/app/cli/internal/cli/command/commands.go`) drive the same `Registry`/`Settings`/`Syncer` from the CLI.

### Persistence

No ORM — each feature's repo lives under `internal/service/repository/<name>/` as a plain Go interface + domain types, with a `postgres/` subpackage implementing it on `pgx/v5` (`*pgxpool.Pool`, wired by `internal/infra/postgres`). `postgres/query.sql` in each of those subpackages is hand-written and is the source of truth; `postgres/internal/gen` is generated entirely by `sqlc` (config: `sqlc.yaml` at the repo root, schema source `migrations/postgres`) — never hand-edit files under `internal/gen`, run `sqlc generate` instead. Repo/translator/disgo interfaces used in tests are mocked the same way via `mockery` (config: `.mockery.yml`) into sibling `mock`/`*mock` packages.

Migrations are plain SQL files under `migrations/postgres/`, managed by `goose` (`github.com/pressly/goose/v3`) — create new ones with `just create-migration <name>`. They are **not** applied automatically by any app process; see the `cli migrate` commands above.

`internal/infra/redis` is also wired into the bot (config: the `redis:` YAML key, supporting multiple named connections each with their own DB index via `Manager.Connection(name)`) but as of now nothing in `internal/service` consumes it — it's only health-checked (a `Ping`) at startup. Treat it as available infra, not as evidence that some feature is Redis-backed.

### i18n

`internal/infra/translator` wraps `go-i18n`; per-locale JSON message files live in `i18n/`. Slash command responses are localized via the `Translator` interface rather than hardcoded strings.

### Deployment

`.github/workflows/test.yml` runs on every push/PR (lint + test against live Postgres/Lavalink containers). `.github/workflows/deploy.yml` additionally builds and pushes the bot Docker image and deploys to Kubernetes via Kustomize (`k8s/base` + `k8s/overlays/prod`) on pushes to `main` (skipped for tag pushes, per commit `cbcbdd9`).

Migrations and Discord command registration are applied by a dedicated `discord-bot-predeploy` Kubernetes Job (`k8s/base/bot/predeploy-job.yaml`, running `./cli migrate up --store postgres && ./cli commands sync`) rather than by the bot process itself — this keeps every bot replica/shard from re-registering commands for its guilds on every restart; the running bot only reacts to a single guild joining (`GuildJoin`) or a guild toggling a command via the `/commands list|enable|disable` subcommands. The workflow pauses the `discord-bot` Deployment's rollout first, so the new image is never rolled out before this job has run, then recreates the Job (Job specs are immutable, so the previous one is deleted first) and waits for it to complete. If it fails, the workflow fails with the Job's logs and the rollout stays paused (old pods keep serving) until a fix is deployed; on success it resumes the rollout and waits for it to finish. (On a namespace's very first-ever deploy there's nothing to pause yet, so that one run doesn't get this ordering guarantee.)
