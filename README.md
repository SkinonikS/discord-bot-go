# Discord Bot

A Discord bot written in Go with temporary voice channel management, reaction-based role assignment, auto roles, and music playback via Lavalink.
Built with [Uber Fx](https://github.com/uber-go/fx) for dependency injection.

Beyond running as-is, the project is designed to double as a **boilerplate for your own bots**: infrastructure (`internal/infra/*`) and features (`internal/service/*`) are isolated, self-contained Fx modules that register themselves into shared value groups (event listeners, slash commands, cron jobs, readiness checks) instead of being wired up centrally. Dropping the existing feature modules and adding new ones following the same shape (`fx.go` + `service.go` + a `*_command.go`/`discord_event_listener.go`) is enough to get a bot with different functionality — see `CLAUDE.md` for the module pattern to copy.

## Features

- **Temporary Voice Channels**: When a user joins a designated "root" voice channel, the bot automatically creates a personal channel for them in a configured category. The channel is deleted once it becomes empty.
- **Reaction Roles**: Assign a role to members when they react to a specific message with a specific emoji; remove the role when the reaction is removed.
- **Auto Roles**: Automatically assign one or more configured roles to members as soon as they join the server.
- **Music Player**: Play audio in voice channels via [Lavalink](https://github.com/lavalink-devs/Lavalink) with queue support.
- **Per-guild command management**: Guild admins can enable/disable individual feature slash commands for their server.
- Localized command responses (English/Russian) via `go-i18n`.

## Requirements

- Go 1.27+
- PostgreSQL
- Redis
- A Discord application with a bot token ([Discord Developer Portal](https://discord.com/developers/applications))
- A running Lavalink node (for music playback)

## Setup Development Environment

1. Copy `.env.example` to `.env` and fill in the required values:

```env
APP_NAME=discord-bot
APP_DEBUG=true
APP_REPOSITORY=https://github.com/your-username/your-repository

DISCORD_TOKEN=your_bot_token_here
DISCORD_APP_ID=your_application_id_here
DISCORD_SHARD_ID=0
DISCORD_SHARD_COUNT=1
DISCORD_WORKER_COUNT=5

DB_POSTGRES_HOST=
DB_POSTGRES_USER=
DB_POSTGRES_PASSWORD=
DB_POSTGRES_DB=
DB_POSTGRES_PORT=5432
DB_POSTGRES_SSLMODE=disable

LAVALINK_MAIN_NODE_HOST=
LAVALINK_MAIN_NODE_PORT=2333
LAVALINK_MAIN_NODE_PASSWORD=
LAVALINK_MAIN_NODE_SECURE=

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

LOG_LEVEL=info
LOG_DISABLE=false
LOG_FORMAT=pretty

MUSIC_PLAYER_IDLE_TIMEOUT=30m
TRANSLATOR_DEFAULT_LOCALE=en-US
```

2. Start the required services via Docker Compose:

```bash
docker compose up -d
```

This starts PostgreSQL, Redis, a Lavalink node, and [LibreDB Studio](https://github.com/libredb/libredb-studio) (a web UI for browsing the database, on `:3000`). See `docker-compose.yml` for defaults (DB: `discord_bot`, password: `root`).

3. Invite the bot to your server

## Running

Apply database migrations and register Discord commands first - the bot does not do this itself:

```bash
go run ./cmd/cli/main.go migrate up --store postgres
go run ./cmd/cli/main.go commands sync
```

Then start the bot:

```bash
# Run directly
go run ./cmd/bot/main.go

# Or build and run
just build-bot
./bot
```

## CLI

A separate CLI tool is available for database management and Discord command registration:

```bash
just build-cli
./cli migrate up --store postgres          # Apply pending migrations for a store
./cli migrate rollback --store postgres    # Roll back the last migration for a store
./cli commands sync                        # Register global + guild commands with Discord
./cli commands list --guild <id>           # List guild-scoped commands and enabled/disabled status for a guild
```

## Development

```bash
go test -v -count=1 ./...          # Run all tests
golangci-lint run                  # Lint
just create-migration <name>       # Create a new SQL migration (goose)
just gen                           # Regenerate sqlc query code and mockery mocks
```

When `APP_DEBUG=true`, `pprof` endpoints are mounted at `/debug/pprof` on the HTTP server (`HTTP_SERVER_HOST`/`HTTP_SERVER_PORT`, default `:8081`).

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
