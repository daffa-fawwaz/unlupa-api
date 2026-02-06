package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Item status constants
const (
	ItemStatusMenghafal  = "menghafal"
	ItemStatusInterval   = "interval"
	ItemStatusFSRSActive = "fsrs_active"
	ItemStatusGraduate   = "graduate"
)

// Graduation threshold in days
const GraduationIntervalDays = 30

type Item struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	OwnerID    uuid.UUID `gorm:"type:uuid;not null;index"`
	SourceType string    `gorm:"type:varchar(20);not null"` // quran | class | personal
	ContentRef string    `gorm:"not null"`

	// Status State Machine
	Status          string     `gorm:"type:varchar(20);not null;default:'menghafal'"` // menghafal | interval | fsrs_active | graduate
	IntervalDays    int        `gorm:"default:0"`                                      // Custom interval days (for interval phase)
	IntervalStartAt *time.Time `gorm:"type:timestamp"`                                 // When interval started
	IntervalEndAt   *time.Time `gorm:"type:timestamp"`                                 // Deadline for interval

	// FSRS Fields (for fsrs_active phase)
	Stability    float64    `gorm:"default:0"`
	Difficulty   float64    `gorm:"default:5.0"`
	LastReviewAt *time.Time `gorm:"type:timestamp"`
	NextReviewAt *time.Time `gorm:"type:timestamp"`

	CreatedAt time.Time
}

func (i *Item) BeforeCreate(tx *gorm.DB) error {
	i.ID = uuid.New()
	if i.Status == "" {
		i.Status = ItemStatusMenghafal
	}
	return nil
}