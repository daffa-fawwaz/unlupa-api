package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ItemState struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	ItemID uuid.UUID `gorm:"type:uuid;uniqueIndex"`

	State string `gorm:"type:varchar(30);index"`
	/*
	   DORMANT
	   ACQUISITION
	   CONSOLIDATION
	   ACTIVE
	   MAINTENANCE
	   GRADUATED
	   ARCHIVED
	*/

	StateEnteredAt time.Time
}

func (is *ItemState) BeforeCreate(tx *gorm.DB) error {
	is.ID = uuid.New()
	return nil
}