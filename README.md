# Discord Bot

A Discord bot written in Go with temporary voice channel management, reaction-based role assignment, and music playback via Lavalink.
Built with [Uber Fx](https://github.com/uber-go/fx) for dependency injection.

## Features

- **Temporary Voice Channels**: When a user joins a designated "root" voice channel, the bot automatically creates a personal channel for them in a configured category. The channel is deleted once it becomes empty.
- **Reaction Roles**: Assign a role to members when they react to a specific message with a specific emoji; remove the role when the reaction is removed.
- **Music Player**: Play audio in voice channels via [Lavalink](https://github.com/lavalink-devs/Lavalink) with queue support.

## Requirements

- Go 1.26+
- PostgreSQL
- A Discord application with a bot token ([Discord Developer Portal](https://discord.com/developers/applications))
- A running Lavalink node (for music playback)
- `libopus-dev` and `libdave` (see [CGO dependencies](#cgo-dependencies))

## Setup

1. Copy `.env.example` to `.env` and fill in the required values:

```env
APP_NAME=discord-bot
APP_DEBUG=false
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
LAVALINK_MAIN_NODE_SECURE=false

LOG_LEVEL=info
LOG_DISABLE=false

MUSIC_PLAYER_IDLE_TIMEOUT=30m
TRANSLATOR_DEFAULT_LOCALE=en-US
```

2. Start the required services via Docker Compose:

```bash
docker compose up -d
```

This starts PostgreSQL and a Lavalink node. See `docker-compose.yml` for defaults (DB: `discord_bot`, password: `root`).

3. Invite the bot to your server with the following gateway intents enabled:
    - Guilds
    - GuildExpressions
    - GuildVoiceStates
    - GuildMessages
    - GuildMessageReactions

## Running

```bash
# Run directly
CGO_ENABLED=1 go run ./cmd/bot/main.go

# Or build and run
make build-bot
./bot
```

Database migrations are applied automatically on startup.

## CLI

A separate CLI tool is available for database management:

```bash
make build-cli
./cli migrate rollback  # Roll back the last migration
```

## Development

```bash
make test              # Run all tests with race detector
golangci-lint run      # Lint
make migration-create  # Create a new SQL migration (goose)
```

When `APP_DEBUG=true`, a pprof HTTP server starts on `:6060` (configurable via `PPROF_ADDR`).

## CGO Dependencies

The bot uses [DAVE](https://github.com/discord/libdave) - Discord's E2E encryption protocol for voice channels - via the [godave](https://github.com/disgoorg/godave) library. This requires CGO and two native libraries:

- **libopus** - audio codec used for voice encoding/decoding
- **libdave** - Discord's native DAVE implementation

### Installing on Linux (Debian/Ubuntu)

```bash
# libopus
apt-get install libopus-dev

# libdave - build from source using the install script bundled with godave
git clone https://github.com/disgoorg/godave
cd godave/scripts
./libdave_install.sh v1.1.1
export PKG_CONFIG_PATH=$HOME/.local/lib/pkgconfig
```

After installing, build with CGO enabled:

```bash
CGO_ENABLED=1 make build-bot
```

### Docker

The provided `Dockerfile` handles all CGO dependencies automatically — no manual setup needed when deploying via Docker.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
