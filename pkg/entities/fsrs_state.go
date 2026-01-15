package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type FSRSState struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	ItemID uuid.UUID `gorm:"type:uuid;uniqueIndex"`

	// ===== FSRS CORE =====
	Stability  float64
	Difficulty float64

	ElapsedDays   int
	ScheduledDays int

	LastReview time.Time
	NextReview time.Time

	Reps   int
	Lapses int

	// ===== FSRS V6 PARAMS SNAPSHOT =====
	// 17 weights (optional snapshot for audit)
	Params datatypes.JSON `gorm:"type:jsonb"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (fs *FSRSState) BeforeCreate(tx *gorm.DB) error {
	fs.ID = uuid.New()
	return nil
}