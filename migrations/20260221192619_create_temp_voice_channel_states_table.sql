-- +goose Up
-- +goose StatementBegin
CREATE TABLE temp_voice_channel_states(
    id UUID PRIMARY KEY,
    channel_id TEXT NOT NULL,
    member_count INTEGER NOT NULL DEFAULT 0,
    UNIQUE(channel_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE temp_voice_channel_states;
-- +goose StatementEnd
