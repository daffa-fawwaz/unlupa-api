package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IntervalReviewLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	ItemID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Rating     int       `gorm:"not null"` // 1=bad, 2=good, 3=perfect
	ReviewedAt time.Time `gorm:"not null"`
	CreatedAt  time.Time
}

func (l *IntervalReviewLog) BeforeCreate(tx *gorm.DB) error {
	l.ID = uuid.New()
	return nil
}
