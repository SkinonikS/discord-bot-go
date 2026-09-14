-- +goose Up
-- +goose StatementBegin
ALTER TABLE reaction_roles ALTER COLUMN guild_id DROP DEFAULT;
ALTER TABLE temp_voice_channels ALTER COLUMN guild_id DROP DEFAULT;
ALTER TABLE temp_voice_channels ALTER COLUMN guild_id DROP DEFAULT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE reaction_roles ALTER COLUMN guild_id SET DEFAULT 0;
ALTER TABLE temp_voice_channels ALTER COLUMN guild_id SET DEFAULT 0;
ALTER TABLE temp_voice_channels ALTER COLUMN guild_id SET DEFAULT 0;
-- +goose StatementEnd