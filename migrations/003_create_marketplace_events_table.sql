-- +migrate Up
CREATE TABLE marketplace_events (
    id UUID DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    owner VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL CHECK (action IN ('purchased', 'cancelled', 'pending_change', 'pending_change_cancelled', 'changed'))
);

-- +migrate Down
DROP TABLE IF EXISTS marketplace_events;
