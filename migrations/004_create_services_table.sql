-- +migrate Up
CREATE TABLE services (
    id UUID DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    environment_id UUID REFERENCES environments(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    image TEXT NOT NULL,
    build TEXT NOT NULL,
    index INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

-- +migrate Down
DROP TABLE IF EXISTS services;
