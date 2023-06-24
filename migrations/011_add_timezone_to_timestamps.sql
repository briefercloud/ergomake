-- +migrate Up
ALTER TABLE environments
    ALTER COLUMN created_at TYPE timestamp with time zone,
    ALTER COLUMN updated_at TYPE timestamp with time zone,
    ALTER COLUMN deleted_at TYPE timestamp with time zone;

ALTER TABLE env_vars
    ALTER COLUMN created_at TYPE timestamp with time zone,
    ALTER COLUMN updated_at TYPE timestamp with time zone;

ALTER TABLE environment_limits
    ALTER COLUMN created_at TYPE timestamp with time zone,
    ALTER COLUMN updated_at TYPE timestamp with time zone,
    ALTER COLUMN deleted_at TYPE timestamp with time zone;

ALTER TABLE marketplace_events
    ALTER COLUMN created_at TYPE timestamp with time zone,
    ALTER COLUMN updated_at TYPE timestamp with time zone,
    ALTER COLUMN deleted_at TYPE timestamp with time zone;

ALTER TABLE private_registries
    ALTER COLUMN created_at TYPE timestamp with time zone,
    ALTER COLUMN updated_at TYPE timestamp with time zone,
    ALTER COLUMN deleted_at TYPE timestamp with time zone;

ALTER TABLE services
    ALTER COLUMN created_at TYPE timestamp with time zone,
    ALTER COLUMN updated_at TYPE timestamp with time zone,
    ALTER COLUMN deleted_at TYPE timestamp with time zone;

ALTER TABLE users
    ALTER COLUMN created_at TYPE timestamp with time zone,
    ALTER COLUMN updated_at TYPE timestamp with time zone,
    ALTER COLUMN deleted_at TYPE timestamp with time zone;

-- +migrate Down
ALTER TABLE environments
    ALTER COLUMN created_at TYPE timestamp without time zone,
    ALTER COLUMN updated_at TYPE timestamp without time zone,
    ALTER COLUMN deleted_at TYPE timestamp without time zone;

ALTER TABLE env_vars
    ALTER COLUMN created_at TYPE timestamp without time zone,
    ALTER COLUMN updated_at TYPE timestamp without time zone,
    ALTER COLUMN deleted_at TYPE timestamp without time zone;

ALTER TABLE environment_limits
    ALTER COLUMN created_at TYPE timestamp without time zone,
    ALTER COLUMN updated_at TYPE timestamp without time zone,
    ALTER COLUMN deleted_at TYPE timestamp without time zone;

ALTER TABLE marketplace_events
    ALTER COLUMN created_at TYPE timestamp without time zone,
    ALTER COLUMN updated_at TYPE timestamp without time zone,
    ALTER COLUMN deleted_at TYPE timestamp without time zone;

ALTER TABLE private_registries
    ALTER COLUMN created_at TYPE timestamp without time zone,
    ALTER COLUMN updated_at TYPE timestamp without time zone,
    ALTER COLUMN deleted_at TYPE timestamp without time zone;

ALTER TABLE services
    ALTER COLUMN created_at TYPE timestamp without time zone,
    ALTER COLUMN updated_at TYPE timestamp without time zone,
    ALTER COLUMN deleted_at TYPE timestamp without time zone;

ALTER TABLE users
    ALTER COLUMN created_at TYPE timestamp without time zone,
    ALTER COLUMN updated_at TYPE timestamp without time zone,
    ALTER COLUMN deleted_at TYPE timestamp without time zone;
