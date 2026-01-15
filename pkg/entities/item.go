package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Item struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	OwnerID   uuid.UUID `gorm:"type:uuid;not null;index"`
	SourceType string    `gorm:"type:varchar(20);not null"` // quran | class | personal
	ContentRef string    `gorm:"not null"`
	CreatedAt  time.Time
}

func (i *Item) BeforeCreate(tx *gorm.DB) error {
	i.ID = uuid.New()
	return nil
}