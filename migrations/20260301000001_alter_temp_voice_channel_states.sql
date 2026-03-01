-- +goose Up
ALTER TABLE temp_voice_channel_states DROP COLUMN member_count;

-- +goose Down
ALTER TABLE temp_voice_channel_states ADD COLUMN member_count INTEGER NOT NULL DEFAULT 0;
