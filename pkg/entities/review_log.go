package entities

import (
	"time"

	"github.com/google/uuid"
)

type ReviewLog struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	ItemID uuid.UUID `gorm:"type:uuid;not null;index"`
	CardID uuid.UUID `gorm:"type:uuid;index"`

	ReviewedAt time.Time `gorm:"not null"`

	Rating int `gorm:"not null"` // 1=Again, 2=Hard, 3=Good, 4=Easy

	// STATE SEBELUM
	StabilityBefore  float64 `gorm:"not null"`
	DifficultyBefore float64 `gorm:"not null"`

	// STATE SESUDAH
	StabilityAfter  float64 `gorm:"not null"`
	DifficultyAfter float64 `gorm:"not null"`

	IntervalDays int `gorm:"not null"`

	CreatedAt time.Time
}
