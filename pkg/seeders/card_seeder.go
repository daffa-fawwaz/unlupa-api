package seeders

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hifzhun-api/pkg/entities"
)

func SeedCards(db *gorm.DB) error {
	now := time.Now()

	ownerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	// Get existing items for this owner
	var items []entities.Item
	if err := db.Where("owner_id = ?", ownerID).Find(&items).Error; err != nil {
		return err
	}

	if len(items) == 0 {
		return fmt.Errorf("no items found for owner %s, run SeedItems first", ownerID)
	}

	for _, item := range items {
		card := entities.Card{
	ID:           uuid.New(),
	OwnerID:      ownerID,
	ItemID:       item.ID,
	Source:       item.SourceType,
	RefID:        item.ContentRef,
	Stability:    0.4,
	Difficulty:   5.0,
	LastReviewAt: time.Time{}, // 🔥 FIX
	CreatedAt:    now,
	UpdatedAt:    now,
}


		if err := db.Create(&card).Error; err != nil {
			return err
		}
	}

	fmt.Printf("✅ Seeded %d cards\n", len(items))
	return nil
}
