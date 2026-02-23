package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Item status constants
const (
	ItemStatusMenghafal       = "menghafal"
	ItemStatusInterval        = "interval"
	ItemStatusFSRSActive      = "fsrs_active"
	ItemStatusPendingGraduate = "pending_graduate"
	ItemStatusGraduate        = "graduate"
	ItemStatusInactive        = "inactive" // For book items only - user can deactivate/reactivate
)

// Graduation threshold in days
const GraduationIntervalDays = 30

// Graduate review interval in days (review every 20 days after graduation)
const GraduateReviewDays = 20

type Item struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	OwnerID    uuid.UUID `gorm:"type:uuid;not null;index"`
	SourceType string    `gorm:"type:varchar(20);not null"` // quran | class | personal
	ContentRef string    `gorm:"not null"`

	// Status State Machine
	Status          string     `gorm:"type:varchar(20);not null;default:'menghafal'"` // menghafal | interval | fsrs_active | pending_graduate | graduate
	IntervalDays         int        `gorm:"default:0"`          // Custom interval days (for interval phase)
	IntervalStartAt      *time.Time `gorm:"type:timestamp"`     // When interval started
	IntervalEndAt        *time.Time `gorm:"type:timestamp"`     // Legacy: deadline for interval (kept for compatibility)
	IntervalNextReviewAt *time.Time `gorm:"type:timestamp"`     // Next recurring interval review date

	// FSRS Fields (for fsrs_active phase)
	Stability    float64    `gorm:"default:0"`
	Difficulty   float64    `gorm:"default:5.0"`
	ReviewCount  int        `gorm:"default:0"`   // Total number of reviews
	LastReviewAt *time.Time `gorm:"type:timestamp"`
	NextReviewAt *time.Time `gorm:"type:timestamp"`

	// Teacher Approval Fields (for pending_graduate phase)
	ApprovedBy *uuid.UUID `gorm:"type:uuid;index"` // Teacher who approved graduation
	ApprovedAt *time.Time `gorm:"type:timestamp"`  // When graduation was approved

	CreatedAt time.Time
}

func (i *Item) BeforeCreate(tx *gorm.DB) error {
	i.ID = uuid.New()
	if i.Status == "" {
		i.Status = ItemStatusMenghafal
	}
	return nil
}