package entities

import (
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

type ReviewState struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	ItemID uuid.UUID `gorm:"type:uuid;not null;index"`
	CardID uuid.UUID `gorm:"type:uuid;index"`

	State string `gorm:"type:varchar(20);not null"` 
	// new | learning | review | maintenance | frozen | graduated
	Source string `gorm:"size:32"`

	Stability  float64 `gorm:"not null"`
	Difficulty float64 `gorm:"not null"`

	LastReviewedAt *time.Time
	NextReviewAt   *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (rs *ReviewState) BeforeCreate(tx *gorm.DB) error {
	rs.ID = uuid.New()
	return nil
}