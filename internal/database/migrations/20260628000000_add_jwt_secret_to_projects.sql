-- +goose Up
-- +goose StatementBegin
ALTER TABLE fluxend.projects ADD COLUMN IF NOT EXISTS jwt_secret TEXT NOT NULL DEFAULT '';
UPDATE fluxend.projects SET jwt_secret = encode(gen_random_bytes(32), 'hex') WHERE jwt_secret = '';
ALTER TABLE fluxend.projects ALTER COLUMN jwt_secret DROP DEFAULT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE fluxend.projects DROP COLUMN IF EXISTS jwt_secret;
-- +goose StatementEnd
