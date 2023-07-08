package database

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"database/sql"
)

type EnvStatus string

const (
	EnvPending  EnvStatus = "pending"
	EnvBuilding EnvStatus = "building"
	EnvSuccess  EnvStatus = "success"
	EnvDegraded EnvStatus = "degraded"
	EnvLimited  EnvStatus = "limited"
	EnvStale    EnvStatus = "stale"
)

type Environment struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	Owner          string
	BranchOwner    string
	Repo           string
	Branch         sql.NullString
	PullRequest    sql.NullInt32
	Author         string
	Status         EnvStatus
	DegradedReason json.RawMessage `gorm:"type:jsonb"`
	Services       []Service       `gorm:"foreignKey:EnvironmentID"`
	GHCommentID    int64           `gorm:"column:gh_comment_id"`
	BuildTool      string
}

func NewEnvironment(
	ID uuid.UUID,
	owner, branchOwner, repo, branch string,
	pullRequest *int, author string, status EnvStatus,
) *Environment {
	pr := 0
	if pullRequest != nil {
		pr = *pullRequest
	}

	return &Environment{
		ID:          ID,
		Owner:       owner,
		BranchOwner: branchOwner,
		Repo:        repo,
		Branch: sql.NullString{
			String: branch,
			Valid:  branch != "",
		},
		PullRequest: sql.NullInt32{
			Int32: int32(pr),
			Valid: pr != 0,
		},
		Author: author,
		Status: status,
	}
}

func (db *DB) FindEnvironmentByID(id uuid.UUID) (Environment, error) {
	var env Environment
	result := db.
		Preload("Services", func(db *gorm.DB) *gorm.DB {
			return db.Order("services.index ASC")
		}).
		First(&env, "id = ?", id)

	return env, result.Error
}

type FindEnvironmentsOptions struct {
	IncludeDeleted bool
}

func (db *DB) FindEnvironmentsByPullRequest(
	pullRequest int,
	owner string,
	repo string,
	branch string,
	options FindEnvironmentsOptions,
) ([]Environment, error) {
	envs := make([]Environment, 0)
	where := map[string]interface{}{
		"pull_request": pullRequest,
		"owner":        owner,
		"repo":         repo,
		"branch":       branch,
	}

	result := db.Where(where)
	if options.IncludeDeleted {
		result = db.Unscoped().Where(where)
	}

	result = result.Order("created_at ASC").
		Preload("Services", func(db *gorm.DB) *gorm.DB {
			return db.Order("services.index ASC")
		}).
		Find(&envs)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return envs, nil
	}

	return envs, result.Error
}

func (db *DB) DeleteEnvironmentByPullRequest(pullRequest int, owner, repo, branch string) error {
	result := db.Where(map[string]interface{}{
		"pull_request": pullRequest,
		"owner":        owner,
		"repo":         repo,
		"branch":       branch,
	}).Delete(&Environment{})

	return result.Error
}

func (db *DB) FindEnvironmentsByOwner(owner string, options FindEnvironmentsOptions) ([]Environment, error) {
	envs := make([]Environment, 0)

	where := map[string]interface{}{
		"owner": owner,
	}

	result := db.DB
	if options.IncludeDeleted {
		result = db.Unscoped()
	}

	result = result.Where(where).Order("created_at DESC").
		Preload("Services", func(db *gorm.DB) *gorm.DB {
			return db.Order("services.index ASC")
		}).
		Find(&envs)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return envs, nil
	}

	return envs, result.Error
}
