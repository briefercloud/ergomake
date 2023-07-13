-- +migrate Up

ALTER TABLE env_vars ADD COLUMN branch VARCHAR(255);

ALTER TABLE env_vars DROP CONSTRAINT env_vars_owner_repo_name_key;
ALTER TABLE env_vars ADD CONSTRAINT env_vars_owner_repo_name_key UNIQUE (owner, repo, name, branch);

-- +migrate Down

ALTER TABLE env_vars DROP COLUMN branch;
