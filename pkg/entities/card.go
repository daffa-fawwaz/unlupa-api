package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Card struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	// Kepemilikan & konteks (FSRS TIDAK PEDULI)
	OwnerID uuid.UUID `gorm:"type:uuid;index"`
	ItemID  uuid.UUID `gorm:"type:uuid;not null;index"`
	Source  string    `gorm:"size:32;index"`
	RefID   string    `gorm:"size:64;index"`

	// FSRS PURE STATE
	Stability  float64 `gorm:"not null"`
	Difficulty float64 `gorm:"not null"`

	LastReviewAt time.Time `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// ✅ GORM HOOK — FIXED
func (c *Card) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
