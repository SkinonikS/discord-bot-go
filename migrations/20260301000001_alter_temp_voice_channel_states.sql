-- +goose Up
-- +goose StatementBegin
ALTER TABLE temp_voice_channel_states DROP COLUMN member_count;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE temp_voice_channel_states ADD COLUMN member_count INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd