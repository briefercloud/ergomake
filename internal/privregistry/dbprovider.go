package privregistry

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/ergomake/ergomake/internal/crypto"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/dockerutils"
)

type privateRegistry struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deletedAt"`
	Owner       string
	URL         string
	Provider    string
	Credentials string
}

type dbPrivRegistryProvider struct {
	db     *database.DB
	secret string
}

func NewDBPrivRegistryProvider(db *database.DB, secret string) *dbPrivRegistryProvider {
	return &dbPrivRegistryProvider{db, secret}
}

func (prp *dbPrivRegistryProvider) FetchCreds(
	ctx context.Context,
	owner string,
	image string,
) (*RegistryCreds, error) {
	url, err := dockerutils.ExtractDockerRegistryURL(image)
	if err != nil {
		return nil, errors.Wrap(err, "fail to extract docker registry URL")
	}

	var registry privateRegistry
	err = prp.db.First(&registry, map[string]string{"owner": owner, "url": url}).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrRegistryNotFound
	}

	if err != nil {
		return nil, errors.Wrapf(err, "fail to find private docker registry credentials in db for url %s", url)
	}

	token, err := prp.getToken(registry)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to get token of registry %s", registry.URL)
	}

	return &RegistryCreds{ID: registry.ID, URL: url, Provider: registry.Provider, Token: token}, nil
}

func (prp *dbPrivRegistryProvider) getToken(registry privateRegistry) (string, error) {
	creds, err := crypto.Decrypt(prp.secret, registry.Credentials)
	if err != nil {
		return "", errors.Wrap(err, "fail to decrypt credentials")
	}

	var token string
	switch registry.Provider {
	case "ecr":
		token, err = prp.getECRToken(registry.URL, creds)
		if err != nil {
			return "", errors.Wrapf(err, "fail to fetch token from ecr for url %s", registry.URL)
		}
	default:
		return "", errors.Errorf("fail to extract token for %s got unexpected provider %s", registry.URL, registry.Provider)
	}

	return token, nil
}

func (prp *dbPrivRegistryProvider) StoreRegistry(
	ctx context.Context,
	owner string,
	url string,
	provider string,
	credentials string,
) error {
	credentials, err := crypto.Encrypt(prp.secret, credentials)
	if err != nil {
		return errors.Wrap(err, "fail to encrypt credentials")
	}

	registry := privateRegistry{
		Owner:       owner,
		URL:         url,
		Provider:    provider,
		Credentials: credentials,
	}
	err = prp.db.Create(&registry).Error
	return errors.Wrap(err, "fail to save registry to db")
}

func (prp *dbPrivRegistryProvider) ListCredsByOwner(ctx context.Context, owner string, skipToken bool) ([]RegistryCreds, error) {
	creds := make([]RegistryCreds, 0)
	var privRegistries []privateRegistry
	err := prp.db.Find(&privRegistries, map[string]string{"owner": owner}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return creds, nil
	}

	for _, pr := range privRegistries {
		var token string
		var err error
		if !skipToken {
			token, err = prp.getToken(pr)
		}

		if err != nil {
			return nil, errors.Wrap(err, "fail to get token")
		}

		creds = append(creds, RegistryCreds{
			ID:       pr.ID,
			URL:      pr.URL,
			Provider: pr.Provider,
			Token:    token,
		})
	}

	return creds, nil
}

func (prp *dbPrivRegistryProvider) DeleteRegistry(ctx context.Context, id uuid.UUID) error {
	registry := privateRegistry{ID: id}

	err := prp.db.Delete(&registry).Error
	if err != nil {
		return errors.Wrapf(err, "fail to delete registry with ID %s from db", id)
	}

	return nil
}

type ecrCredentials struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}

func (prp *dbPrivRegistryProvider) getECRToken(url string, rawCreds string) (string, error) {
	var creds ecrCredentials
	err := json.Unmarshal([]byte(rawCreds), &creds)
	if err != nil {
		return "", errors.Wrapf(err, "fail to decode ecr credentials for %s", url)
	}

	return GetECRToken(url, creds.AccessKeyID, creds.SecretAccessKey, creds.Region)
}
