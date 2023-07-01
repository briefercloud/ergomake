-- +migrate Up
CREATE TABLE stripe_subscriptions (
    id UUID DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    owner VARCHAR(255) NOT NULL,
    subscription_id VARCHAR(255) NOT NULL,
    UNIQUE(owner, subscription_id)
);

-- +migrate Down
DROP TABLE IF EXISTS stripe_subscriptions;
