package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClassBook struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ClassID   uuid.UUID `gorm:"type:uuid;not null;index" json:"class_id"`
	BookID    uuid.UUID `gorm:"type:uuid;not null;index" json:"book_id"`
	Order     int       `gorm:"default:0" json:"order"`
	CreatedAt time.Time `json:"created_at"`

	// Relations
	Book Book `gorm:"foreignKey:BookID" json:"book,omitempty"`
}

func (cb *ClassBook) BeforeCreate(tx *gorm.DB) error {
	cb.ID = uuid.New()
	return nil
}
