-- +migrate Up
DROP TABLE IF EXISTS registries;
CREATE TABLE private_registries (
    id UUID DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    owner VARCHAR(255) NOT NULL,
    url VARCHAR(255) NOT NULL,
    provider VARCHAR(255) NOT NULL CHECK (provider in ('ecr', 'gcr', 'hub')),
    credentials TEXT NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS private_registries;
CREATE TABLE registries (
    id UUID DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    owner VARCHAR(255) NOT NULL,
    url VARCHAR(255) NOT NULL,
    token TEXT NOT NULL
);
