package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ItemGraduation struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	ItemID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Keputusan akhir
	Action string `gorm:"size:16;not null"` 
	// graduate | freeze | reactivate

	Reason string `gorm:"size:255"`

	CreatedAt time.Time
}

func (ig *ItemGraduation) BeforeCreate(tx *gorm.DB) error {
	if ig.ID == uuid.Nil {
		ig.ID = uuid.New()
	}
	return nil
}
