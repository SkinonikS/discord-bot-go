# Discord Bot

A Discord bot written in Go that provides automatic temporary voice channel management and reaction-based role assignment.
Built with [Uber Fx](https://github.com/uber-go/fx) for dependency injection.

## Features

- **Temporary Voice Channels**: When a user joins a designated "root" voice channel, the bot automatically creates a personal channel for them in a configured category. The channel is deleted once it becomes empty.
- **Reaction Roles**: Assign a role to members when they react to a specific message with a specific emoji; remove the role when the reaction is removed.

## Requirements

- Go 1.25.1+
- A Discord application with a bot token ([Discord Developer Portal](https://discord.com/developers/applications))

## Setup

1. Copy `.env.example` to `.env` (or create `.env`) and fill in the required values:

```env
DISCORD_TOKEN=your_bot_token_here
DISCORD_APP_ID=your_application_id_here
DISCORD_APP_OWNER_ID=your_discord_user_id_here

APP_DEBUG=true
APP_REPOSITORY=https://github.com/your-username/discord-bot-go

LOG_LEVEL=info
LOG_DISABLE=false
```

2. Configure `config/config.yaml` if needed (defaults are reasonable for development).

3. Invite the bot to your server with the following intents enabled:
   - **Server Members Intent** (for role assignment)
   - **Message Content Intent** (for reactions)
   - Required permissions:
     - Manage Channels
     - Manage Roles
     - Add Reactions

## Running

```bash
# Run directly
go run ./cmd/bot/main.go

# Or build first
go build -o discord-bot ./cmd/bot
./discord-bot
```

The SQLite database and migrations are handled automatically on startup.
