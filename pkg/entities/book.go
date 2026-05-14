package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Book status constants
const (
	BookStatusDraft     = "draft"
	BookStatusPending   = "pending"
	BookStatusPublished = "published"
	BookStatusRejected  = "rejected"
)

type Book struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OwnerID uuid.UUID `gorm:"type:uuid;not null;index" json:"owner_id"`

	Title       string `gorm:"size:200;not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	CoverImage  string `gorm:"size:500" json:"cover_image"`

	// IsEditable controls whether users who import/copy this published book
	// can add, edit, or delete items and modules on their own copy.
	// true  = importers can freely edit their copy
	// false = importers cannot modify items/modules (read-only for them)
	IsEditable bool `gorm:"not null;default:true" json:"is_editable"`

	Status      string     `gorm:"size:20;not null;default:'draft'" json:"status"`
	PublishedAt *time.Time `json:"published_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Modules []BookModule `gorm:"foreignKey:BookID" json:"modules,omitempty"`
	Items   []BookItem   `gorm:"foreignKey:BookID" json:"items,omitempty"`
}

func (b *Book) BeforeCreate(tx *gorm.DB) error {
	b.ID = uuid.New()
	if b.Status == "" {
		b.Status = BookStatusDraft
	}
	return nil
}
