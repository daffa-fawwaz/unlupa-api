package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Class struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	GuruID uuid.UUID `gorm:"type:uuid;not null;index" json:"guru_id"`
	Name      string `gorm:"size:100;not null" json:"name"`
	ClassCode string `gorm:"size:20;uniqueIndex;not null" json:"class_code"`
	CreatedAt time.Time `json:"created_at"`
}

func (c *Class) BeforeCreate(tx *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}