package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DailyTask struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	// Kepemilikan
	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	ItemID uuid.UUID `gorm:"type:uuid;not null;index"`
	CardID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Snapshot date (WAJIB)
	TaskDate time.Time `gorm:"type:date;not null;index"`

	// Metadata
	Source string `gorm:"size:32;not null"` // quran | kitab | personal

	// Status task harian
	State string `gorm:"size:16;not null"` // pending | done | skipped

	CreatedAt time.Time
}

func (dt *DailyTask) BeforeCreate(tx *gorm.DB) error {
	if dt.ID == uuid.Nil {
		dt.ID = uuid.New()
	}
	return nil
}
