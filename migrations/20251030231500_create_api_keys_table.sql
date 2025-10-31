-- +goose Up
CREATE TABLE IF NOT EXISTS api_keys (
    id SERIAL PRIMARY KEY,
    key_hash TEXT NOT NULL UNIQUE,
    client_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_keys_client_name ON api_keys (client_name);

INSERT INTO api_keys (key_hash, client_name)
VALUES ('8ac950188678f9bb3524b275130332b511bf5092394da6975b5fb9e84302f026', 'test_client');
-- raw key (do not store in DB): test-client-secret

-- +goose Down
DROP TABLE IF EXISTS api_keys;
