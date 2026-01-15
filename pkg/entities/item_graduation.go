package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ItemGraduation struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	ItemID uuid.UUID `gorm:"type:uuid;uniqueIndex"`

	Reason string
	// USER_DECISION | SYSTEM_RULE

	FrozenAt time.Time
	ReactivatedAt *time.Time
}

func (ig *ItemGraduation) BeforeCreate(tx *gorm.DB) error {
	ig.ID = uuid.New()
	return nil
}
