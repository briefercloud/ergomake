package envvars

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/ergomake/ergomake/internal/crypto"
	"github.com/ergomake/ergomake/internal/database"
)

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type EnvVarsProvider interface {
	Upsert(ctx context.Context, owner, repo, name, value string) error
	Delete(ctx context.Context, owner, repo, name string) error
	ListByRepo(ctx context.Context, owner, repo string) ([]EnvVar, error)
}

type DBEnvVar struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Owner     string
	Repo      string
	Name      string
	Value     string
}

type dbEnvVarsProvider struct {
	db     *database.DB
	secret string
}

func NewDBEnvVarProvider(db *database.DB, secret string) *dbEnvVarsProvider {
	return &dbEnvVarsProvider{db, secret}
}

func (evp *dbEnvVarsProvider) Upsert(ctx context.Context, owner, repo, name, value string) error {
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

func (evp *dbEnvVarsProvider) Delete(ctx context.Context, owner, repo, name string) error {
	err := evp.db.Table("env_vars").Where(map[string]interface{}{
		"owner": owner,
		"repo":  repo,
		"name":  name,
	}).Delete(&DBEnvVar{}).Error

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

		vars = append(vars, EnvVar{v.Name, value})
	}

	return vars, err
}
