-- +migrate Up
CREATE TABLE deployed_branches (
    id UUID DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    owner VARCHAR(255) NOT NULL,
    repo VARCHAR(255) NOT NULL,
    branch VARCHAR(255) NOT NULL,
    UNIQUE(owner, repo, branch)
);

-- +migrate Down
DROP TABLE IF EXISTS deployed_branches;
