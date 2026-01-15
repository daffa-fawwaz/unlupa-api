package seeders

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hifzhun-api/pkg/entities"
)

func SeedItems(db *gorm.DB, ownerID uuid.UUID) error {
	items := []entities.Item{
		{
			OwnerID:   ownerID,
			SourceType: "quran",
			ContentRef: "Al-Baqarah:1",
		},
		{
			OwnerID:   ownerID,
			SourceType: "quran",
			ContentRef: "Al-Baqarah:2",
		},
		{
			OwnerID:   ownerID,
			SourceType: "personal",
			ContentRef: "Mufradat Bab 1",
		},
	}

	for _, item := range items {
		var count int64

		err := db.Model(&entities.Item{}).
			Where("owner_id = ? AND source_type = ? AND content_ref = ?",
				item.OwnerID, item.SourceType, item.ContentRef).
			Count(&count).Error

		if err != nil {
			return err
		}

		if count > 0 {
			continue // sudah ada → skip
		}

		if err := db.Create(&item).Error; err != nil {
			return err
		}

		fmt.Println("seeded item:", item.ContentRef)
	}

	return nil
}
