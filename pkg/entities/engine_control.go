package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EngineControl struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	ItemID uuid.UUID `gorm:"type:uuid;not null;index"`

	IsFrozen     bool
	FrozenReason *string
	FrozenAt     *time.Time

	IsGraduated bool
	GraduatedAt *time.Time

	CreatedAt time.Time
}

func (ec *EngineControl) BeforeCreate(tx *gorm.DB) error {
	ec.ID = uuid.New()
	return nil
}