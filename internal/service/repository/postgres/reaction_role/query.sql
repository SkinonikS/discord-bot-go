-- name: FindReactionRolesByCriteria :many
SELECT *
FROM reaction_roles
WHERE (guild_id = sqlc.narg(guild_id) OR sqlc.narg(guild_id) = 0)
  AND (channel_id = sqlc.narg(channel_id) OR sqlc.narg(channel_id) = 0)
  AND (message_id = sqlc.narg(message_id) OR sqlc.narg(message_id) = 0)
  AND (emoji_name = sqlc.narg(emoji_name) OR sqlc.narg(emoji_name) = '')
  AND (role_id = sqlc.narg(role_id) OR sqlc.narg(role_id) = 0);

-- name: SaveReactionRole :one
INSERT INTO reaction_roles (id, guild_id, channel_id, message_id, emoji_name, role_id)
VALUES (@id, @guild_id, @channel_id, @message_id, @emoji_name, @role_id)
    ON CONFLICT (guild_id, message_id, emoji_name)
DO UPDATE SET role_id = EXCLUDED.role_id
           RETURNING id;

-- name: DeleteReactionRolesByIDs :execrows
DELETE FROM reaction_roles
WHERE id = ANY(@ids::uuid[]);
