package entities

import (
	"github.com/google/uuid"
)

type JuzItem struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	JuzID  uuid.UUID `gorm:"type:uuid;not null;index"`
	ItemID uuid.UUID `gorm:"type:uuid;not null;index"`
}
