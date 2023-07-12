package permanentbranches

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/ergomake/ergomake/internal/database"
)

type deployedBranch struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Owner     string         `gorm:"index"`
	Repo      string         `gorm:"index"`
	Branch    string         `gorm:"index"`
}

type dbPermanentBranchesProvider struct {
	db *database.DB
}

func NewDBEnvironmentsProvider(
	db *database.DB,
) *dbPermanentBranchesProvider {
	return &dbPermanentBranchesProvider{db}
}

func (ep *dbPermanentBranchesProvider) List(ctx context.Context, owner, repo string) ([]string, error) {
	deployedBranches := make([]*deployedBranch, 0)
	branches := make([]string, 0)

	err := ep.db.Table("deployed_branches").
		Find(&deployedBranches, map[string]string{
			"owner": owner,
			"repo":  repo,
		}).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return branches, nil
		}

		return branches, errors.Wrap(err, "fail to query deployed_branches table")
	}

	for _, deployedBranch := range deployedBranches {
		branches = append(branches, deployedBranch.Branch)
	}

	return branches, nil
}

func (ep *dbPermanentBranchesProvider) IsPermanentBranch(ctx context.Context, owner, repo, branch string) (bool, error) {
	var deployedBranch deployedBranch
	err := ep.db.Table("deployed_branches").First(
		&deployedBranch,
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

// BatchUpsert updates the branches in the database and returns the result of the operation
func (pbp *dbPermanentBranchesProvider) BatchUpsert(ctx context.Context, owner, repo string, branches []string) (BatchUpsertResult, error) {
	res := BatchUpsertResult{
		Added:   make([]string, 0),
		Removed: make([]string, 0),
		Result:  branches,
	}

	tx := pbp.db.Debug().Begin()
	if tx.Error != nil {
		return res, errors.Wrap(tx.Error, "transaction failed to begin")
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Fetch all current branches (including soft-deleted ones) before any operations are started
	var initialBranches []deployedBranch
	dbChain := tx.Table("deployed_branches").
		Where("owner = ? AND repo = ?", owner, repo).
		Find(&initialBranches)

	if err := dbChain.Error; err != nil {
		return res, errors.Wrap(err, "failed to fetch current branches")
	}

	for _, branch := range branches {
		var deployedBranch deployedBranch
		deployedBranch.Owner = owner
		deployedBranch.Repo = repo
		deployedBranch.Branch = branch

		// Fetch from database, including soft deleted entities
		dbChain := tx.Unscoped().Table("deployed_branches").
			Where(deployedBranch).
			First(&deployedBranch)

		added := false

		// Either create new record, or "undelete" if it was soft-deleted
		if errors.Is(dbChain.Error, gorm.ErrRecordNotFound) {
			dbChain = tx.Table("deployed_branches").Create(&deployedBranch)
			added = true
		} else if deployedBranch.DeletedAt.Valid {
			deployedBranch.DeletedAt = gorm.DeletedAt{Time: time.Time{}}
			dbChain = tx.Table("deployed_branches").Save(&deployedBranch)
			added = true
		}

		if err := dbChain.Error; err != nil {
			tx.Rollback()
			return res, errors.Wrapf(err, "failed to upsert permanent branch %s to %s/%s", branch, owner, repo)
		}

		if added {
			res.Added = append(res.Added, branch)
		}
	}

	if len(branches) > 0 {
		dbChain = tx.Table("deployed_branches").
			Where("owner = ? AND repo = ?", owner, repo).
			Not("branch", branches).
			Delete(&deployedBranch{})
	} else {
		dbChain = tx.Table("deployed_branches").
			Where("owner = ? AND repo = ?", owner, repo).
			Delete(&deployedBranch{})
	}

	if err := dbChain.Error; err != nil {
		tx.Rollback()
		return res, errors.Wrap(err, "failed to delete branches in the transaction")
	}

	// Call getRemoved with initialBranches and the final state of branches
	res.Removed = getRemoved(initialBranches, branches)

	if err := tx.Commit().Error; err != nil {
		return res, errors.Wrap(err, "transaction failed to commit")
	}

	return res, nil
}

// getRemoved is a helper function to find which branches were removed
func getRemoved(initialBranches []deployedBranch, finalBranches []string) []string {
	removed := []string{}
	mapping := make(map[string]bool)

	// Map finalBranches for easy look up
	for _, branch := range finalBranches {
		mapping[branch] = true
	}

	// Check each branch in initialBranches in the map to spot the missing ones
	for _, initialBranch := range initialBranches {
		if !mapping[initialBranch.Branch] {
			removed = append(removed, initialBranch.Branch)
		}
	}

	return removed
}
