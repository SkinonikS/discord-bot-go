-- +goose Up
-- +goose StatementBegin
CREATE TABLE reaction_roles (
    id UUID PRIMARY KEY,
    channel_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    emoji_name TEXT NOT NULL,
    role_id TEXT NOT NULL,
    UNIQUE (message_id, emoji_name)
);

CREATE INDEX idx_reaction_roles_message_id ON reaction_roles(message_id);
CREATE INDEX idx_reaction_roles_role_id ON reaction_roles(role_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_reaction_roles_message_id;
DROP INDEX IF EXISTS idx_reaction_roles_role_id;
DROP TABLE IF EXISTS reaction_roles;
-- +goose StatementEnd
