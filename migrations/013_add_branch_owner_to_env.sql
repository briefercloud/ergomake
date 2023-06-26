-- +migrate Up
ALTER TABLE environments ADD COLUMN branch_owner VARCHAR(255) NOT NULL DEFAULT '';
UPDATE environments SET branch_owner = owner;
ALTER TABLE environments ALTER COLUMN branch_owner DROP DEFAULT;

-- +migrate Down
ALTER TABLE environments DROP COLUMN branch_owner;
