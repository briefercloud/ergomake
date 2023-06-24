-- +migrate Up
ALTER TABLE environments DROP CONSTRAINT environments_status_check;

ALTER TABLE environments
ADD CONSTRAINT environments_status_check
CHECK (status IN ('pending', 'building', 'success', 'degraded', 'limited', 'stale'));

-- +migrate Down
UPDATE environments SET status = 'degraded' WHERE status = 'stale';

ALTER TABLE DROP CONSTRAINT environments_status_check;

ALTER TABLE environments
ADD CONSTRAINT environments_status_check
CHECK (status IN ('pending', 'building', 'success', 'degraded', 'limited'));
