-- +migrate Up
CREATE TABLE environment_limits (
    id UUID DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    owner VARCHAR(255) NOT NULL,
    env_limit INT NOT NULL,
    UNIQUE(owner)
);

-- +migrate Down
DROP TABLE IF EXISTIS environment_limits;
