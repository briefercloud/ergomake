-- +migrate Up

ALTER TABLE environments ADD COLUMN build_tool VARCHAR(255) NOT NULL DEFAULT 'kaniko';

-- +migrate Down

ALTER TABLE environments DROP COLUMN build_tool;
