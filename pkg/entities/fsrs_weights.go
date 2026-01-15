package entities

import (
	"time"

	"github.com/google/uuid"
)

type FSRSWeights struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	OwnerID uuid.UUID `gorm:"type:uuid;index"` // per user / global

	Version string `gorm:"size:8;index"` // "v6"

	W0  float64
	W1  float64
	W2  float64
	W3  float64
	W4  float64
	W5  float64
	W6  float64
	W7  float64
	W8  float64
	W9  float64
	W10 float64
	W11 float64
	W12 float64
	W13 float64
	W14 float64
	W15 float64
	W16 float64

	CreatedAt time.Time
}
