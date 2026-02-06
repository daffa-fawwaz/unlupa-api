package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookModule struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	BookID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"book_id"`
	ParentID *uuid.UUID `gorm:"type:uuid;index" json:"parent_id,omitempty"` // untuk nested module (opsional)

	Title       string `gorm:"size:200;not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	Order       int    `gorm:"not null;default:0" json:"order"`

	CreatedAt time.Time `json:"created_at"`

	// Relations
	Items []BookItem `gorm:"foreignKey:ModuleID" json:"items,omitempty"`
}

func (m *BookModule) BeforeCreate(tx *gorm.DB) error {
	m.ID = uuid.New()
	return nil
}
