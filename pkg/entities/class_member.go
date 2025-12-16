package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClassMember struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	ClassID uuid.UUID `gorm:"type:uuid;not null;index" json:"class_id"`
	UserID  uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`

	JoinedAt time.Time `json:"joined_at"`
}

func (cm *ClassMember) BeforeCreate(tx *gorm.DB) error {
	cm.ID = uuid.New()
	return nil
}