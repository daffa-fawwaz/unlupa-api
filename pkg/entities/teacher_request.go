package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeacherRequest struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	Message   string    `gorm:"type:text" json:"message"`
	Status    string    `gorm:"size:20;not null;default:pending" json:"status"` // pending, approved, rejected
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (t *TeacherRequest) BeforeCreate(tx *gorm.DB) error {
	t.ID = uuid.New()
	return nil
}
