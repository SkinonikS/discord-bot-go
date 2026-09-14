-- +goose Up
-- +goose StatementBegin
CREATE TABLE disabled_guild_commands (
    id UUID PRIMARY KEY,
    guild_id BIGINT NOT NULL,
    command_name TEXT NOT NULL,
    UNIQUE (guild_id, command_name)
);

CREATE INDEX idx_disabled_guild_commands_guild_shard ON disabled_guild_commands ((guild_id >> 22));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_disabled_guild_commands_guild_shard;
DROP TABLE IF EXISTS disabled_guild_commands;
-- +goose StatementEnd
