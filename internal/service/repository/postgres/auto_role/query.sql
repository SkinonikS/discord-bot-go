-- name: FindAutoRolesByGuildID :many
SELECT *
FROM auto_roles
WHERE guild_id = @guild_id;

-- name: SaveAutoRole :exec
INSERT INTO auto_roles (id, guild_id, role_id)
VALUES (@id, @guild_id, @role_id)
    ON CONFLICT (guild_id, role_id) DO NOTHING;

-- name: DeleteAutoRoleByGuildIDAndRoleID :execrows
DELETE FROM auto_roles
WHERE guild_id = @guild_id AND role_id = @role_id;
