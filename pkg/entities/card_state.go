package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CardState struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	CardID uuid.UUID `gorm:"type:uuid;not null;index" json:"card_id"`

	Stability  float64 `gorm:"not null" json:"stability"`
	Difficulty float64 `gorm:"not null" json:"difficulty"`

	LastReviewedAt *time.Time `json:"last_reviewed_at,omitempty"`
	NextReviewAt   *time.Time `json:"next_review_at,omitempty"`

	ReviewCount int `json:"review_count"`
	LapseCount  int `json:"lapse_count"`

	Status string `gorm:"size:10;not null" json:"status"`
}

func (cs *CardState) BeforeCreate(tx *gorm.DB) error {
	cs.ID = uuid.New()
	return nil
}