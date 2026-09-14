-- name: IsGuildCommandDisabled :one
SELECT EXISTS(
    SELECT 1
    FROM disabled_guild_commands
    WHERE guild_id = @guild_id AND command_name = @command_name
);

-- name: ListDisabledGuildCommandsByGuildID :many
SELECT command_name
FROM disabled_guild_commands
WHERE guild_id = @guild_id;

-- name: DisableGuildCommand :exec
INSERT INTO disabled_guild_commands (id, guild_id, command_name)
VALUES (@id, @guild_id, @command_name)
    ON CONFLICT (guild_id, command_name) DO NOTHING;

-- name: EnableGuildCommand :exec
DELETE FROM disabled_guild_commands
WHERE guild_id = @guild_id AND command_name = @command_name;
