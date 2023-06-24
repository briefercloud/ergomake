-- +migrate Up
ALTER TABLE environments ADD COLUMN gh_comment_id BIGINT NOT NULL DEFAULT 0;

-- +migrate Down
ALTER TABLE environments DROP COLUMN gh_comment_id;
