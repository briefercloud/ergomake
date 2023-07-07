-- +migrate Up

ALTER TABLE services
ADD COLUMN public_port VARCHAR(255),
ADD COLUMN internal_ports text[],
ADD COLUMN build_status VARCHAR(255)
CHECK (build_status IN ('image', 'building', 'build-failed', 'build-success'));

-- +migrate Down

ALTER TABLE services
DROP COLUMN public_port,
DROP COLUMN internal_ports,
DROP COLUMN build_status;
