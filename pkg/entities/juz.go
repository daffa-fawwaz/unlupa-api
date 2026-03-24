package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Juz struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	Index  int       `gorm:"not null"` // 1 - 30

	IsActive bool `gorm:"default:true"`
	IsDone   bool `gorm:"default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DoneAt    *time.Time
}

func (j *Juz) BeforeCreate(tx *gorm.DB) error {
	j.ID = uuid.New()
	return nil
}
