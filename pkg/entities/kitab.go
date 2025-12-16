package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)


type Kitab struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	OwnerID *uuid.UUID `gorm:"type:uuid;index" json:"owner_id,omitempty"`
	ClassID *uuid.UUID `gorm:"type:uuid;index" json:"class_id,omitempty"`

	Title       string `gorm:"size:150;not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	Type        string `gorm:"size:10;not null" json:"type"` // QURAN | CLASS | PERSONAL
	IsShared    bool   `gorm:"default:false" json:"is_shared"`

	CreatedAt time.Time `json:"created_at"`
}

func (k *Kitab) BeforeCreate(tx *gorm.DB) error {
	k.ID = uuid.New()
	return nil
}
