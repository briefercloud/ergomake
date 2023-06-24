-- +migrate Up
CREATE TABLE environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    owner VARCHAR(255) NOT NULL,
    repo VARCHAR(255) NOT NULL,
    branch VARCHAR(255),
    pull_request INT,
    author VARCHAR(255),
    status VARCHAR(255) NOT NULL CHECK (status IN ('pending', 'building', 'success', 'degraded', 'limited'))
);

-- +migrate Down
DROP TABLE environments;
