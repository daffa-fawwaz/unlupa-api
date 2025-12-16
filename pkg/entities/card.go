package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Card struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	KitabID uuid.UUID `gorm:"type:uuid;not null;index" json:"kitab_id"`

	OrderIndex int    `gorm:"not null" json:"order_index"`
	ContentText string `gorm:"type:text" json:"content_text"`

	ContentAudioURL string `gorm:"size:255" json:"content_audio_url"`
	ContentImageURL string `gorm:"size:255" json:"content_image_url"`

	CreatedAt time.Time `json:"created_at"`
}
func (c *Card) BeforeCreate(tx *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}