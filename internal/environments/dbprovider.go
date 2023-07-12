package environments

import (
	"context"
	"time"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/ergomake/ergomake/internal/cluster"
	"github.com/ergomake/ergomake/internal/database"
	"github.com/ergomake/ergomake/internal/payment"
	"github.com/ergomake/ergomake/internal/permanentbranches"
)

type environmentLimits struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Owner     string         `gorm:"index"`
	EnvLimit  int
}

type dbEnvironmentsProvider struct {
	db                        *database.DB
	paymentProvider           payment.PaymentProvider
	envLimitAmount            int
	permanentBranchesProvider permanentbranches.PermanentBranchesProvider
	clusterClient             cluster.Client
}

func NewDBEnvironmentsProvider(
	db *database.DB,
	paymentProvider payment.PaymentProvider,
	envLimitAmount int,
	permanentBranchesProvider permanentbranches.PermanentBranchesProvider,
	clusterClient cluster.Client,
) *dbEnvironmentsProvider {
	return &dbEnvironmentsProvider{db, paymentProvider, envLimitAmount, permanentBranchesProvider, clusterClient}
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

	ownerEnvs, err := ep.db.FindEnvironmentsByOwner(owner, database.FindEnvironmentsOptions{})
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
	isPermanentBranch, err := ep.permanentBranchesProvider.IsPermanentBranch(ctx, owner, repo, branch)

	return isPermanentBranch, errors.Wrapf(err, "fail to check if branch %s is configured as permanent for repo %s/%s", branch, owner, repo)
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

func (ep *dbEnvironmentsProvider) TerminateEnvironment(ctx context.Context, req TerminateEnvironmentRequest) error {
	branchEnvs, err := ep.ListEnvironmentsByBranch(ctx, req.Owner, req.Repo, req.Branch)
	if err != nil {
		return errors.Wrap(err, "fail to list environments by branch")
	}

	envs := make([]*database.Environment, 0)
	for _, env := range branchEnvs {
		if req.PrNumber != nil {
			if env.PullRequest.Valid && env.PullRequest.Int32 == int32(*req.PrNumber) {
				envs = append(envs, env)
			}
		} else {
			envs = append(envs, env)
		}
	}

	for _, env := range envs {
		err = ep.clusterClient.DeleteNamespace(ctx, env.ID.String())
		if err != nil && !k8sErrors.IsNotFound(err) {
			return errors.Wrap(err, "fail to delete namespace")
		}

		err = ep.DeleteEnvironment(ctx, env.ID)
		if err != nil {
			return errors.Wrap(err, "fail to delete environment in DB")
		}
	}

	return nil
}
