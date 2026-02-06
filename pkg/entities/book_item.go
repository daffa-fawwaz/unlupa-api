package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookItem struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	BookID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"book_id"`
	ModuleID *uuid.UUID `gorm:"type:uuid;index" json:"module_id,omitempty"` // null jika langsung di book

	Title   string `gorm:"size:200;not null" json:"title"`
	Content string `gorm:"type:text" json:"content"` // materi konten
	Order   int    `gorm:"not null;default:0" json:"order"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (i *BookItem) BeforeCreate(tx *gorm.DB) error {
	i.ID = uuid.New()
	return nil
}
