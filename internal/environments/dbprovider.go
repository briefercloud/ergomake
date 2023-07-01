package environments

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/payment"
)

type environmentLimits struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Owner     string         `gorm:"index"`
	EnvLimit  int
}

type deployedBranch struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Owner     string         `gorm:"index"`
	Repo      string         `gorm:"index"`
	Branch    string         `gorm:"index"`
}

type dbEnvironmentsProvider struct {
	db              *database.DB
	paymentProvider payment.PaymentProvider
	envLimitAmount  int
}

func NewDBEnvironmentsProvider(
	db *database.DB,
	paymentProvider payment.PaymentProvider,
	envLimitAmount int,
) *dbEnvironmentsProvider {
	return &dbEnvironmentsProvider{db, paymentProvider, envLimitAmount}
}

func (ep *dbEnvironmentsProvider) IsOwnerLimited(ctx context.Context, owner string) (bool, error) {
	limit := ep.envLimitAmount

	var dbOwnerLimit environmentLimits
	err := ep.db.First(&dbOwnerLimit, map[string]string{"owner": owner}).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		plan, err := ep.paymentProvider.GetOwnerPlan(ctx, owner)
		if err != nil {
			return false, errors.Wrapf(err, "fail to get owner %s plan ", owner)
		}

		if plan == payment.PaymentPlanProfessional {
			return false, nil
		}

		if plan == payment.PaymentPlanStandard {
			limit = payment.StandardPlanEnvLimit
		}
	} else if err != nil {
		return false, errors.Wrapf(err, "fail to fetch specific limit configuration for owner %s from db", owner)
	} else {
		limit = dbOwnerLimit.EnvLimit
	}

	ownerEnvs, err := ep.db.FindEnvironmentsByOwner(owner)
	if err != nil {
		return false, errors.Wrapf(err, "fail to get current environments for owner %s", owner)
	}
	currentEnvCount := 0
	for _, env := range ownerEnvs {
		if env.Status == database.EnvLimited {
			continue
		}
		currentEnvCount += 1
	}

	return currentEnvCount >= limit, nil
}

func (ep *dbEnvironmentsProvider) GetEnvironmentFromHost(
	ctx context.Context,
	host string,
) (*database.Environment, error) {
	var env database.Environment
	err := ep.db.Table("environments").Select("environments.*").
		Joins("INNER JOIN services s ON s.environment_id = environments.id").
		Where("s.url = ?", host).
		Order("environments.created_at DESC").
		Preload("Services").
		First(&env).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Wrapf(ErrEnvironmentNotFound, "no environmnet found for host %s", host)
	}

	return &env, errors.Wrapf(err, "fail to query for environment of host %s", host)
}

func (ep *dbEnvironmentsProvider) SaveEnvironment(ctx context.Context, env *database.Environment) error {
	err := ep.db.Save(env).Error
	return errors.Wrap(err, "fail to save environment to db")
}

func (ep *dbEnvironmentsProvider) ListSuccessEnvironments(ctx context.Context) ([]*database.Environment, error) {
	environments := make([]*database.Environment, 0)
	err := ep.db.Table("environments").Where("status = ?", database.EnvSuccess).
		Preload("Services").Find(&environments).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return environments, nil
		}

		return nil, errors.Wrap(err, "failed to query for success environments")
	}

	return environments, nil
}

func (ep *dbEnvironmentsProvider) ShouldDeploy(ctx context.Context, owner string, repo string, branch string) (bool, error) {
	plan, err := ep.paymentProvider.GetOwnerPlan(ctx, owner)
	if err != nil {
		return false, errors.Wrapf(err, "fail to get owner %s plan ", owner)
	}

	if plan == payment.PaymentPlanFree {
		return false, nil
	}

	var deployedBranch *deployedBranch
	err = ep.db.Table("deployed_branches").First(
		deployedBranch,
		map[string]string{
			"owner":  owner,
			"repo":   repo,
			"branch": branch,
		},
	).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (ep *dbEnvironmentsProvider) ListEnvironmentsByBranch(
	ctx context.Context,
	owner, repo, branch string,
) ([]*database.Environment, error) {
	envs := make([]*database.Environment, 0)

	err := ep.db.Table("environments").
		Preload("Services", func(db *gorm.DB) *gorm.DB {
			return db.Order("services.index ASC")
		}).
		Find(&envs, map[string]string{
			"owner":  owner,
			"repo":   repo,
			"branch": branch,
		}).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return envs, nil
		}
	}

	return envs, err
}

func (ep *dbEnvironmentsProvider) DeleteEnvironment(ctx context.Context, id uuid.UUID) error {
	return ep.db.Table("environments").Delete(&database.Environment{ID: id}).Error
}
