-- name: FindTempVoiceChannelStatesByCriteria :many
SELECT *
FROM temp_voice_channel_states
WHERE (guild_id = sqlc.narg(guild_id) OR sqlc.narg(guild_id) = 0)
  AND (channel_id = sqlc.narg(channel_id) OR sqlc.narg(channel_id) = 0);

-- name: SaveTempVoiceChannelState :exec
INSERT INTO temp_voice_channel_states (id, guild_id, channel_id)
VALUES (@id, @guild_id, @channel_id);

-- name: DeleteTempVoiceChannelStatesByIDs :execrows
DELETE FROM temp_voice_channel_states
WHERE id = ANY(@ids::uuid[]);
