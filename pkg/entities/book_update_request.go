package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	BookUpdateStatusPending  = "pending"
	BookUpdateStatusApproved = "approved"
	BookUpdateStatusRejected = "rejected"
)

// BookUpdateRequest tracks update requests for published books
type BookUpdateRequest struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	BookID      uuid.UUID `gorm:"type:uuid;not null;index" json:"book_id"`
	OwnerID     uuid.UUID `gorm:"type:uuid;not null;index" json:"owner_id"`
	Title       string    `gorm:"size:200" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	CoverImage  string    `gorm:"size:500" json:"cover_image"`
	Status      string    `gorm:"size:20;not null;default:'pending'" json:"status"`
	RequestedAt time.Time `json:"requested_at"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`
	ApprovedBy  *uuid.UUID `gorm:"type:uuid" json:"approved_by,omitempty"`
	RejectReason string   `gorm:"type:text" json:"reject_reason,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	Book   *Book   `gorm:"foreignKey:BookID" json:"book,omitempty"`
	Owner  *User   `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
}

func (b *BookUpdateRequest) BeforeCreate(tx *gorm.DB) error {
	b.ID = uuid.New()
	if b.Status == "" {
		b.Status = BookUpdateStatusPending
	}
	b.RequestedAt = time.Now()
	return nil
}
