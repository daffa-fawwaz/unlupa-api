package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Juz struct {
	ID      uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ClassID *uuid.UUID `gorm:"type:uuid;index" json:"class_id,omitempty"`
	Index   int        `gorm:"not null" json:"index"` // 1 - 30

	IsActive bool `gorm:"default:true" json:"is_active"`
	IsDone   bool `gorm:"default:false" json:"is_done"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DoneAt    *time.Time `json:"done_at,omitempty"`
}

func (j *Juz) BeforeCreate(tx *gorm.DB) error {
	j.ID = uuid.New()
	return nil
}
