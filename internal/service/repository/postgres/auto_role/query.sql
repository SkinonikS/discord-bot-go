-- name: FindAutoRoles :many
SELECT *
FROM auto_roles
WHERE (guild_id = @guild_id OR @guild_id = 0)
    OR (role_id = @role_id OR @role_id = 0);

-- name: SaveAutoRole :exec
INSERT INTO auto_roles (id, guild_id, role_id)
VALUES (@id, @guild_id, @role_id)
    ON CONFLICT (guild_id, role_id) DO NOTHING;

-- name: DeleteAutoRoles :execrows
DELETE FROM auto_roles
WHERE (guild_id = @guild_id OR @guild_id = 0)
  AND (role_id = @role_id OR @role_id = 0);
