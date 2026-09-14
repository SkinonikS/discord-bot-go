-- +goose Up
-- +goose StatementBegin
CREATE TABLE auto_roles (
    id UUID PRIMARY KEY,
    guild_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    UNIQUE (guild_id, role_id)
);

CREATE INDEX idx_auto_roles_guild_shard ON auto_roles ((guild_id >> 22));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_auto_roles_guild_shard;
DROP TABLE IF EXISTS auto_roles;
-- +goose StatementEnd
