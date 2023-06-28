-- +migrate Up

ALTER TABLE environments ADD COLUMN degraded_reason JSONB;

-- +migrate Down

ALTER TABLE environments DROP COLUMN degraded_reason;
