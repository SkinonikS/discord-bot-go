-- +goose Up
-- +goose StatementBegin
ALTER TABLE temp_voice_channels
    ADD COLUMN guild_id BIGINT DEFAULT 0,
    ALTER COLUMN root_channel_id TYPE BIGINT USING root_channel_id::BIGINT,
    ALTER COLUMN parent_id TYPE BIGINT USING parent_id::BIGINT,
    DROP CONSTRAINT temp_voice_channels_root_channel_id_key,
    ADD CONSTRAINT temp_voice_channels_guild_id_root_channel_id_key UNIQUE (guild_id, root_channel_id);

ALTER TABLE temp_voice_channel_states
    ADD COLUMN guild_id BIGINT DEFAULT 0,
    ALTER COLUMN channel_id TYPE BIGINT USING channel_id::BIGINT,
    DROP CONSTRAINT temp_voice_channel_states_channel_id_key,
    ADD CONSTRAINT temp_voice_channel_states_guild_id_channel_id_key UNIQUE (guild_id, channel_id);

ALTER TABLE reaction_roles
    ADD COLUMN guild_id BIGINT DEFAULT 0,
    ALTER COLUMN channel_id TYPE BIGINT USING channel_id::BIGINT,
    ALTER COLUMN message_id TYPE BIGINT USING message_id::BIGINT,
    ALTER COLUMN role_id TYPE BIGINT USING role_id::BIGINT,
    DROP CONSTRAINT reaction_roles_message_id_emoji_name_key,
    ADD CONSTRAINT reaction_roles_guild_id_message_id_emoji_name_key UNIQUE (guild_id, message_id, emoji_name);

CREATE INDEX idx_temp_voice_channels_guild_shard ON temp_voice_channels ((guild_id >> 22));
CREATE INDEX idx_temp_voice_channel_states_guild_shard ON temp_voice_channel_states ((guild_id >> 22));
CREATE INDEX idx_reaction_roles_guild_shard ON reaction_roles ((guild_id >> 22));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE temp_voice_channels
    DROP COLUMN IF EXISTS guild_id,
    ALTER COLUMN root_channel_id TYPE TEXT USING root_channel_id::TEXT,
    ALTER COLUMN parent_id TYPE TEXT USING parent_id::TEXT,
    DROP CONSTRAINT IF EXISTS temp_voice_channels_guild_id_root_channel_id_key,
    ADD CONSTRAINT temp_voice_channels_root_channel_id_key UNIQUE (root_channel_id);

ALTER TABLE temp_voice_channel_states
    DROP COLUMN IF EXISTS guild_id,
    ALTER COLUMN channel_id TYPE TEXT USING channel_id::TEXT,
    DROP CONSTRAINT IF EXISTS temp_voice_channel_states_guild_id_channel_id_key,
    ADD CONSTRAINT temp_voice_channel_states_channel_id_key UNIQUE (channel_id);

ALTER TABLE reaction_roles
    DROP COLUMN IF EXISTS guild_id,
    ALTER COLUMN channel_id TYPE TEXT USING channel_id::TEXT,
    ALTER COLUMN message_id TYPE TEXT USING message_id::TEXT,
    ALTER COLUMN role_id TYPE TEXT USING role_id::TEXT,
    DROP CONSTRAINT IF EXISTS reaction_roles_guild_id_message_id_emoji_name_key,
    ADD CONSTRAINT reaction_roles_message_id_emoji_name_key UNIQUE (message_id, emoji_name);

DROP INDEX IF EXISTS idx_temp_voice_channels_guild_shard;
DROP INDEX IF EXISTS idx_temp_voice_channel_states_guild_shard;
DROP INDEX IF EXISTS idx_reaction_roles_guild_shard;
-- +goose StatementEnd