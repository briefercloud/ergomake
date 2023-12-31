package database

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Service struct {
	ID            string `gorm:"primaryKey"`
	Name          string
	EnvironmentID uuid.UUID `gorm:"type:uuid;index"`
	Url           string
	Image         string
	Build         string
	BuildStatus   string
	Index         int
	PublicPort    string
	InternalPorts pq.StringArray `gorm:"type:text[]"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (db *DB) FindServicesByEnvironment(environmentID uuid.UUID) ([]Service, error) {
	services := make([]Service, 0)
	result := db.Where(map[string]interface{}{
		"environment_id": environmentID,
	}).Order("index ASC").Find(&services)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return services, nil
	}

	return services, result.Error
}
