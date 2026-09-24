# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

Build via `just` (not `make` — there is no Makefile):

```bash
just build-bot               # build ./bot binary (cmd/bot/main.go)
just build-cli                # build ./cli binary (cmd/cli/main.go)
just create-migration name    # generate a new goose SQL migration
just gen                      # sqlc generate + mockery (see "Code generation" below)
```

**Gotcha:** `create-migration` runs `goose -dir migrations create <name> sql`, which writes the new file directly under `migrations/`, not `migrations/postgres/` where every existing migration actually lives and where both the postgres migrator (`internal/infra/foundation.Path.MigrationsPath("postgres")`) and `sqlc.yaml`'s schema source point. After running it, move the generated file into `migrations/postgres/` yourself (or pass `-dir migrations/postgres` instead) before it will be picked up by `cli migrate up` or `sqlc generate`.

Run directly:

```bash
CGO_ENABLED=1 go run ./cmd/bot/main.go
CGO_ENABLED=1 go run ./cmd/cli/main.go migrate up --store postgres        # apply pending migrations for a store
CGO_ENABLED=1 go run ./cmd/cli/main.go migrate rollback --store postgres  # roll back the last migration for a store
CGO_ENABLED=1 go run ./cmd/cli/main.go commands sync                      # register global + guild commands with Discord
CGO_ENABLED=1 go run ./cmd/cli/main.go commands list --guild <id>         # list guild-scoped commands and enabled/disabled status for a guild
```

Test and lint (CI in `.github/workflows/test.yml` / `deploy.yml` runs both against live Postgres + Lavalink service containers):

```bash
go test -v -race -count=1 ./...   # run all tests
go test -race -run TestName ./internal/service/auto_role/...  # single test
golangci-lint run                  # lint (config: .golangci.yml)
```

The only `*_test.go` files in the repo today are the feature `service_test.go`s (`auto_role`, `reaction_role`, `temp_voice_channel`, `music_player`) — all pure unit tests driving the service against `mockery`-generated mocks (`.../repository/*/mock`, `internal/infra/discord/mock`, `internal/service/music_player/mock`). None of them open a real Postgres connection or talk to Lavalink, so `DB_POSTGRES_*`/`LAVALINK_*` env vars aren't actually required to run `go test ./...` locally; CI provisions the containers regardless (presumably for future integration tests, or just to mirror runtime requirements).

CGO is required (the bot links `libopus` and `libdave` for Discord's DAVE voice encryption via `github.com/disgoorg/godave`). See README.md's "CGO Dependencies" section for install steps; Docker builds handle this automatically.

Setting `APP_DEBUG=true` mounts `pprof` at `/debug/pprof` on the main HTTP server (`gofiber/fiber/v3/middleware/pprof`, `HTTP_SERVER_HOST`/`HTTP_SERVER_PORT`, default `:8081`) — there is no separate pprof port.

The bot does **not** apply migrations or register commands itself. Both are applied with the `cli` binary, one store at a time for migrations — `cli migrate up --store <name>` applies pending migrations, `cli migrate rollback --store <name>` rolls one back (both default `--store` to `postgres`, currently the only registered store) — and account-wide for commands via `cli commands sync`. In the deploy pipeline these run as a dedicated Kubernetes Job (see "Deployment" below) before the bot rolls out; locally, run them yourself after `just create-migration` or before starting the bot for the first time.

### Code generation

Repository query code and test mocks are generated, not hand-written — after changing a `query.sql` file or a migration's schema, or an interface that's mocked, regenerate rather than editing generated output by hand:

```bash
sqlc generate   # regenerates internal/service/repository/postgres/*/internal/gen from query.sql + migrations/postgres (config: sqlc.yaml)
mockery         # regenerates */mock packages from the interfaces listed in .mockery.yml
```

## Architecture

This is a Discord bot built entirely around **Uber Fx** dependency injection. There is no code that isn't wired through an `fx.Module` — when adding anything new, the pattern below is not optional style, it's how the app boots.

### Layering: `internal/infra` vs `internal/service`

- **`internal/infra/*`** — infrastructure/platform modules (`config`, `logger`, `postgres`, `redis`, `discord`, `lavalink`, `migrator`, `cron`, `cache`, `translator`, `http_server`, `readiness`, `cli`, `foundation`). Each is an isolated Go package with its own `fx.go` exposing `NewModule() fx.Option`, plus a `config.go` reading its own YAML section via `go.uber.org/config` (where it has runtime config at all).
- **`internal/service/*`** — feature/business modules: `auto_role`, `reaction_role`, `temp_voice_channel`, `music_player`, `interaction_command`, plus two tiny system-command modules, `info` and `ping`, which exist only to register their slash command. Every one of these depends on `interaction_command` (to register a `Command`) and/or `internal/infra/discord` (to register an event listener), but never on another feature module directly.
- Go package names drop the directory's underscores (idiomatic Go) — this trips up naive greps: `auto_role` → package `autorole`, `reaction_role` → `reactionrole`, `temp_voice_channel` → `tempvoicechannel`, `music_player` → `musicplayer`, `interaction_command` → `interactioncommand`, and the `disabled_guild_commands` repo → package `guildcommandsetting`. Search by directory path, not package name.

### Two binaries wire an (almost) identical Fx graph

`internal/app/bot/app.go` and `internal/app/cli/app.go` both build `fx.New(infra.NewModule(params), service.NewModule(), ...)` — the exact same infra + feature module set, differing only in the `foundation.RunMode` supplied (`RunModeApplication` vs `RunModeCLI`), the `cli` app additionally decorating the logger to `zap.NewNop()` and doing `fx.Populate(&cmd)` to pull out the root `*cli.Command`.

This matters because of how each `main.go` drives the resulting `*fx.App`:

- `cmd/bot/main.go` calls `app.Run()`, which starts the Fx **lifecycle** — every `fx.Lifecycle`-registered `OnStart` hook fires (opening the Discord gateway, starting the HTTP server listener, starting the worker pool, starting the cron scheduler, etc.).
- `cmd/cli/main.go` never calls `app.Start()`/`app.Run()` at all — it only checks `app.Err()` (did the graph build?) then runs the populated `urfave/cli` command tree directly. **Lifecycle hooks never fire in the `cli` binary.** Eager `fx.Provide`/`fx.Invoke` construction still happens (e.g. a real `*pgxpool.Pool` is opened, a real `disgobot.Client` is constructed, the cron scheduler is built), but nothing that's gated behind an `OnStart` hook — no gateway connection, no HTTP listener, no worker pool goroutines, no cron ticking — ever runs for `cli`.

**Gotcha (unchanged in spirit):** the consumers of a value group are wired via `fx.Invoke` (e.g. `discord.NewModule()`'s worker pool/`*disgobot.Client`, `lavalink.NewModule()`'s `disgolink.Client`, `cron.NewModule()`'s `gocron.Scheduler`, `http_server.NewModule()`'s `*fiber.App`), and those invokes run for both binaries during graph construction — so any binary wiring `service.NewModule()` transitively needs every dependency that every registered command/listener needs to construct, even if that binary (currently `cli`) never uses it at runtime. E.g. `music_player`'s slash command needs `disgolink.Client`, so `lavalink.NewModule()` must stay wired into `cli` even though `cli` never plays audio. Forgetting this surfaces as an fx "missing type" error at the group's `fx.Invoke` site, not at the module that actually needs the dependency.

**Current state:** `cron.NewModule()` *is* wired into the shared `infra.NewModule()` graph (`internal/infra/fx.go`) and its scheduler is invoked eagerly, but no feature module currently registers anything via `cron.AsJob` — the `cron_jobs` value group is empty, so the scheduler starts (in the bot binary; see above) with zero jobs. There is no stale-entry or orphaned-channel cleanup job anywhere in the tree right now; `temp_voice_channel`'s empty-channel deletion and `reaction_role`'s role add/remove are purely event-driven (`discord_event_listener.go`), not scheduled.

### The `cli_commands` value group (`internal/infra/cli`)

`internal/infra/cli` (module name `"cli"`) is the generic building block for the `urfave/cli` command tree, replacing what used to be a bespoke `internal/app/cli/internal/cli` layer: it exposes `AsCommand(f any) any` (wraps a constructor with `fx.ResultTags(`group:"cli_commands"`)`) and `New(Params) *cli.Command`, which builds the root command (named after `config.Config.Name`, with shell-completion enabled) from every `*cli.Command` in that group. Individual subcommands provide themselves into the group from their own module — `internal/infra/migrator` (`NewCLIMigrateCommand`, the `migrate` subcommand) and `internal/service/interaction_command` (`NewCLICommandsCommand` in `cli_commands_command.go`, the `commands` subcommand) both call `cli.AsCommand(...)`. Adding a new CLI subcommand means providing it the same way from wherever it logically belongs — nothing in `internal/infra/cli` itself needs to change.

### The per-module fx.go pattern

Every module (infra or service) follows this shape — copy it when adding a new one:

```go
const ModuleName = "..."

func NewModule() fx.Option {
    return fx.Module(ModuleName,
        fx.Provide(newConfig, NewService, ...),
        fx.Provide(
            discord.AsEventListener(NewDiscordEventListener),
            interactioncommand.AsCommand(NewDiscordFooCommand),
            cron.AsJob(NewFooCronJob),
        ),
        fx.Decorate(func(log *zap.Logger) *zap.Logger {
            return log.Named(ModuleName)
        }),
    )
}
```

The `AsX(f any) any` helper pattern (`discord.AsEventListener`, `interactioncommand.AsCommand`, `cli.AsCommand`, `cron.AsJob`, `readiness.AsHandler`, `httpserver.AsHandler`, `migrator.AsProvider`, `cache.AsDriverFactory`) wraps a constructor with `fx.Annotate(..., fx.As(new(Interface)), fx.ResultTags(`group:"..."`))` so it self-registers into a value group. Consumers collect the group with an `fx.In`-embedded params struct field tagged `group:"..."` (see `interactioncommand.RegistryParams`, `readiness.HandlerParams`). This is how event listeners, slash commands, CLI subcommands, readiness checks, and migration providers are discovered without any central registration list — to add a new one, just provide it with the right `AsX` wrapper from your module's `fx.go`, nothing else needs to change.

### Config

Config is layered YAML + env expansion, not plain env vars:

- `config/config.yaml` (or `config/config.<env>.yaml`) has `${VAR:default}` placeholders expanded against the process environment (`internal/infra/config/config.go`, using `go.uber.org/config`).
- `.env` (or `.env.<env>`) is loaded first via `godotenv` to populate those env vars — see `.env.example` for the full variable list.
- Each infra module reads its own top-level YAML key into its own `*Config` struct via `provider.Get(ConfigKey).Populate(...)` (e.g. `internal/infra/postgres/config.go` reads the `postgres:` key). Follow this pattern for any new configurable module rather than reading `config.yaml` directly.

### Feature module shape (e.g. `internal/service/auto_role`)

A typical feature module has: `service.go` (business logic interface + impl, depends on a `Repo` and possibly the Discord REST client — obtained as `disgobot.Client.Rest`/`RestClient`, part of `internal/infra/discord`'s `Result`, not a separate module), `discord_event_listener.go` (implements `disgobot.EventListener`, registered via `discord.AsEventListener`), and a `discord_*_command.go` (implements `interactioncommand.Command`, registered via `interactioncommand.AsCommand`). All Discord-facing files in a feature module are prefixed `discord_` by convention (`discord_event_listener.go`, `discord_auto_role_command.go`, `discord_music_command.go`, ...).

### Slash commands: registry, per-guild settings, and sync

`internal/service/interaction_command` has these collaborating pieces:

- **`Registry`** (`registry.go`) just collects every `Command` from the `discord_commands` group (populated eagerly via `fx.Invoke(populateRegistry)`).
- **`Command.Scope()`** is either `CommandScopeGuild` (feature commands — can be toggled off per guild) or `CommandScopeGlobal` (system commands like `ping`/`info`/the `commands` management command itself — always registered account-wide, can't be disabled per guild).
- **`Service`** (`service.go`) is both the settings store and the Discord syncer in one: it wraps the `guildcommandsetting` (`disabled_guild_commands` table) repo directly — `IsGuildCommandDisabled`/`ListDisabledGuildCommands`/`SetGuildCommandState` — and pushes definitions to Discord over REST — `SyncGlobalCommands` (all `CommandScopeGlobal` commands), `SyncGuildCommands(guildID)` (that guild's enabled `CommandScopeGuild` commands), `SyncAllGuildCommands` (paginates every guild the bot is currently in via `GetCurrentUserGuildsPage`).

The `commands` slash command (`discord_manage_command.go`, admin-only, guild context) exposes `list`/`enable`/`disable` subcommands backed directly by `Service`, so a server admin can turn a feature off without redeploying. The bot's `discord_event_listener.go` also calls `Service.SyncGuildCommands` on `GuildJoin`. The `cli commands sync` / `cli commands list --guild <id>` subcommands (`cli_commands_command.go`) drive the same `Registry`/`Service` from the CLI.

### HTTP server and readiness

`internal/infra/http_server` wraps [Fiber v3](https://github.com/gofiber/fiber) (not gin — migrated from `gin.go` to `fiber.go`), with request-ID, panic-recovery, and zap-request-logging middleware always on, plus `pprof` when `APP_DEBUG=true`. Route registration is decentralized the same way as everything else: anything implementing `httpserver.Handler` (`Register(app *fiber.App) error`) and provided via `httpserver.AsHandler(...)` (group `http_handlers`) gets wired in.

`internal/infra/readiness` is one level of indirection on top of that: it has its *own* `readiness.Handler` interface (`IsReady()`/`IsHealthy()`) and its own value group (`readiness_handlers`, wired via `readiness.AsHandler(...)` — currently only `internal/infra/discord`'s gateway-status check), aggregated by `readiness.NewRegistry`. That aggregate registry is then wrapped by `NewHTTPHandler` and exposed to the outside world as a single `httpserver.Handler` (provided via `httpserver.AsHandler`) serving `/livez` and `/readyz`. To add a new readiness check, provide a `readiness.Handler` with `readiness.AsHandler(...)` from your module — you don't touch the HTTP layer at all.

### Persistence

No ORM — each feature's repo lives under `internal/service/repository/<name>/` as a plain Go interface + domain types, with a sibling `internal/service/repository/postgres/<name>/` package implementing it on `pgx/v5` (`*pgxpool.Pool`, wired by `internal/infra/postgres`). `postgres/<name>/query.sql` is hand-written and is the source of truth; `postgres/internal/gen` is generated entirely by `sqlc` (config: `sqlc.yaml` at the repo root, schema source `migrations/postgres`) — never hand-edit files under `internal/gen`, run `sqlc generate` instead. `internal/service/repository/fx.go` is the single place binding every Postgres repo implementation to its interface via `fx.Annotate(pgsqlX.NewRepo, fx.As(new(x.Repo)))`. Repo/translator/disgo interfaces used in tests are mocked the same way via `mockery` (config: `.mockery.yml`) into sibling `mock`/`*mock` packages.

Postgres migrations register themselves the same value-group way as everything else: `internal/infra/postgres` provides a `migrator.Provider` (a named `*goose.Provider`) via `migrator.AsProvider(newMigratorProvider)` into the `migrator_providers` group; `internal/infra/migrator.NewRegistry` collects it, and `cli migrate up|rollback --store <name>` looks it up by name (`postgres` today, the only registered store). Migrations are plain SQL files under `migrations/postgres/`, managed by `goose` (`github.com/pressly/goose/v3`) — see the `just create-migration` gotcha above for where new files actually land. They are **not** applied automatically by any app process; see the `cli migrate` commands above.

`internal/infra/redis` is also wired into both binaries (config: the `redis:` YAML key, supporting multiple named connections each with their own DB index via `Manager.Connection(name)`) but as of now nothing in `internal/service` consumes it — it's only health-checked (a `Ping`, logged, not exposed via the `readiness` registry) at startup. Treat it as available infra, not as evidence that some feature is Redis-backed.

### i18n

`internal/infra/translator` wraps `go-i18n`; per-locale JSON message files live in `i18n/`. Slash command responses are localized via the `Translator` interface rather than hardcoded strings.

### Deployment

`.github/workflows/test.yml` runs on every push (except to `main`) and PR (lint + test against live Postgres/Lavalink containers). `.github/workflows/deploy.yml` runs the same test job plus builds/pushes the bot Docker image (containing both the `bot` and `cli` binaries) on pushes to `main` and on `v*` tags, and additionally deploys to Kubernetes via Kustomize (`k8s/base` + `k8s/overlays/prod`) — but only when `github.ref_type != 'tag'`, i.e. tag pushes build and push the image but skip the Kubernetes deploy step.

Migrations and Discord command registration are applied by a dedicated `discord-bot-predeploy` Kubernetes Job (`k8s/base/bot/predeploy-job.yaml`, running `./cli migrate up --store postgres && ./cli commands sync`) rather than by the bot process itself — this keeps every bot replica/shard from re-registering commands for its guilds on every restart; the running bot only reacts to a single guild joining (`GuildJoin`) or a guild toggling a command via the `/commands list|enable|disable` subcommands. The workflow pauses the `discord-bot` Deployment's rollout first, so the new image is never rolled out before this job has run, then recreates the Job (Job specs are immutable, so the previous one is deleted first) and waits for it to complete. If it fails, the workflow fails with the Job's logs and the rollout stays paused (old pods keep serving) until a fix is deployed; on success it resumes the rollout and waits for it to finish. (On a namespace's very first-ever deploy there's nothing to pause yet, so that one run doesn't get this ordering guarantee.)
