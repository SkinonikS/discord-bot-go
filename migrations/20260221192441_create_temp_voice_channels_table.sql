-- +goose Up
-- +goose StatementBegin
CREATE TABLE temp_voice_channels(
    id UUID PRIMARY KEY,
    root_channel_id TEXT NOT NULL,
    parent_id TEXT NOT NULL,
    UNIQUE(root_channel_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE temp_voice_channels;
-- +goose StatementEnd
