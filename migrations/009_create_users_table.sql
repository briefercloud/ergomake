-- +migrate Up
CREATE TABLE users (
    id UUID DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(255) NOT NULL CHECK (provider in ('github', 'bitbucket', 'gitlab')),
    UNIQUE(email, username, provider)
);

-- +migrate Down
DROP TABLE IF EXISTS users;
