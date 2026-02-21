-- +goose Up
-- +goose StatementBegin
CREATE TABLE reaction_roles (
    id UUID PRIMARY KEY,
    channel_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    emoji_name TEXT NOT NULL,
    role_id TEXT NOT NULL,
    UNIQUE(message_id, emoji_name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE reaction_roles;
-- +goose StatementEnd
