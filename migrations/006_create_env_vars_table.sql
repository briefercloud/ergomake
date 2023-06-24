-- +migrate Up
CREATE TABLE env_vars (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    owner VARCHAR(255) NOT NULL,
    repo VARCHAR(255) NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    UNIQUE(owner, repo, name)
);


-- +migrate Down
DROP TABLE env_vars;
