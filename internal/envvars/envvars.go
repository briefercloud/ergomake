package envvars

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"k8s.io/utils/pointer"

	"github.com/ergomake/ergomake/internal/crypto"
	"github.com/ergomake/ergomake/internal/database"
)

type EnvVar struct {
	Name   string  `json:"name"`
	Value  string  `json:"value"`
	Branch *string `json:"branch"`
}

type EnvVarsProvider interface {
	Upsert(ctx context.Context, owner, repo, name, value string, branch *string) error
	Delete(ctx context.Context, owner, repo, name string, branch *string) error
	ListByRepo(ctx context.Context, owner, repo string) ([]EnvVar, error)
	ListByRepoBranch(ctx context.Context, owner, repo, branch string) ([]EnvVar, error)
}

type DBEnvVar struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Owner     string
	Repo      string
	Name      string
	Value     string
	Branch    sql.NullString
}

type dbEnvVarsProvider struct {
	db     *database.DB
	secret string
}

func NewDBEnvVarProvider(db *database.DB, secret string) *dbEnvVarsProvider {
	return &dbEnvVarsProvider{db, secret}
}

func (evp *dbEnvVarsProvider) Upsert(ctx context.Context, owner, repo, name, value string, branch *string) error {
	encryptedValue, err := crypto.Encrypt(evp.secret, value)
	if err != nil {
		return errors.Wrap(err, "failed to encrypt value")
	}

	var dbVar DBEnvVar
	err = evp.db.Table("env_vars").Where(map[string]interface{}{
		"owner": owner,
		"repo":  repo,
		"name":  name,
	}).Assign(map[string]interface{}{
		"value": encryptedValue,
	}).FirstOrCreate(&dbVar).Error

	return errors.Wrap(err, "failed to upsert env var")
}

func (evp *dbEnvVarsProvider) Delete(ctx context.Context, owner, repo, name string, branch *string) error {
	err := evp.db.Table("env_vars").
		Where(map[string]interface{}{
			"owner":  owner,
			"repo":   repo,
			"name":   name,
			"branch": branch,
		}).
		Delete(&DBEnvVar{}).Error

	return err
}

func (evp *dbEnvVarsProvider) ListByRepo(ctx context.Context, owner, repo string) ([]EnvVar, error) {
	vars := make([]EnvVar, 0)
	var dbVars []DBEnvVar
	err := evp.db.Table("env_vars").Where(map[string]string{
		"owner": owner,
		"repo":  repo,
	}).Find(&dbVars).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return vars, nil
	}

	for _, v := range dbVars {
		value, err := crypto.Decrypt(evp.secret, v.Value)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to decrypt value of env var %s", v.ID)
		}

		branch := new(string)
		if v.Branch.Valid {
			branch = pointer.String(v.Branch.String)
		}

		vars = append(vars, EnvVar{v.Name, value, branch})
	}

	return vars, err
}

func (evp *dbEnvVarsProvider) ListByRepoBranch(ctx context.Context, owner, repo, branch string) ([]EnvVar, error) {
	allRepoVars, err := evp.ListByRepo(ctx, owner, repo)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to list env vars for repo %s/%s", owner, repo)
	}

	vars := make(map[string]EnvVar)
	for _, v := range allRepoVars {
		if v.Branch != nil {
			if *v.Branch == branch {
				vars[v.Name] = v
				continue
			}
		}

		_, ok := vars[v.Name]
		if !ok {
			vars[v.Name] = v
		}
	}

	result := make([]EnvVar, len(vars))
	for _, v := range vars {
		result = append(result, v)
	}

	return result, nil
}
