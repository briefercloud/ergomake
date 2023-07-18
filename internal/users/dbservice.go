package users

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/ergomake/ergomake/internal/database"
)

type dbUsersService struct {
	db *database.DB
}

type databaseUser struct {
	User

	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

func NewDBUsersService(db *database.DB) *dbUsersService {
	return &dbUsersService{db}
}

func (up *dbUsersService) Save(ctx context.Context, user User) error {
	var dbUser databaseUser

	var err error
	if user.Email == "" {
		err = up.db.Table("users").
			Find(
				&dbUser,
				"provider = ? AND username = ?",
				user.Provider,
				user.Username,
			).Error
	} else {
		err = up.db.Table("users").
			Find(
				&dbUser,
				"provider = ? AND (email = ? OR username = ?)",
				user.Provider,
				user.Email,
				user.Username,
			).Error
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Wrap(err, "fail to check for existing user in db")
	}

	dbUser.User = user
	err = up.db.Table("users").Save(&dbUser).Error
	return errors.Wrap(err, "fail to save user to db")
}
