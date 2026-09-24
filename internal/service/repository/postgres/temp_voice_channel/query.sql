-- name: FindTempVoiceChannels :many
SELECT *
FROM temp_voice_channels
WHERE (guild_id = sqlc.narg(guild_id) OR sqlc.narg(guild_id) = 0)
  AND (root_channel_id = sqlc.narg(root_channel_id) OR sqlc.narg(root_channel_id) = 0)
  AND (parent_id = sqlc.narg(parent_id) OR sqlc.narg(parent_id) = 0);

-- name: SaveTempVoiceChannel :exec
INSERT INTO temp_voice_channels (id, guild_id, root_channel_id, parent_id)
VALUES (@id, @guild_id, @root_channel_id, @parent_id)
    ON CONFLICT (root_channel_id, guild_id)
DO UPDATE SET parent_id = EXCLUDED.parent_id;

-- name: DeleteTempVoiceChannelsByIDs :execrows
DELETE FROM temp_voice_channels
WHERE id = ANY(@ids::uuid[]);