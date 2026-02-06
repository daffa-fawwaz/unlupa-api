package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Class type constants
const (
	ClassTypeQuran = "quran"
	ClassTypeBook  = "book"
)

type Class struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	GuruID      uuid.UUID `gorm:"type:uuid;not null;index" json:"guru_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	ClassCode   string    `gorm:"size:20;uniqueIndex;not null" json:"class_code"`
	Type        string    `gorm:"size:20;not null;default:'quran'" json:"type"` // quran | book
	IsActive    bool      `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Members []ClassMember `gorm:"foreignKey:ClassID" json:"members,omitempty"`
	Books   []ClassBook   `gorm:"foreignKey:ClassID" json:"books,omitempty"`
}

func (c *Class) BeforeCreate(tx *gorm.DB) error {
	c.ID = uuid.New()
	if c.Type == "" {
		c.Type = ClassTypeQuran
	}
	return nil
}