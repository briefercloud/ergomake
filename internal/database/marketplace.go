package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Action can only be "purchased", "cancelled", "pending_change", "pending_change_cancelled", "changed"
type MarketplaceEvent struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Owner     string
	Action    string
}

func (e *MarketplaceEvent) BeforeCreate(tx *gorm.DB) error {
	e.ID = uuid.NewString()
	return nil
}

func (db *DB) SaveEvent(owner, action string) error {
	event := &MarketplaceEvent{
		Owner:  owner,
		Action: action,
	}

	result := db.Create(event)
	return result.Error
}
