package seeders

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hifzhun-api/pkg/entities"
)

func SeedFSRSWeights(db *gorm.DB) error {
	ownerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	weights := entities.FSRSWeights{
		ID:      uuid.New(),
		OwnerID: ownerID,
		Version: "v6",
		W0:      0.4072,
		W1:      1.1829,
		W2:      3.1262,
		W3:      15.4722,
		W4:      7.2102,
		W5:      0.5316,
		W6:      1.0651,
		W7:      0.0234,
		W8:      1.616,
		W9:      0.1544,
		W10:    1.0824,
		W11:    1.9813,
		W12:    0.0953,
		W13:    0.2975,
		W14:    2.2042,
		W15:    0.2407,
		W16:    2.9466,
		CreatedAt: time.Now(),
	}

	if err := db.Create(&weights).Error; err != nil {
		return err
	}

	fmt.Println("✅ Seeded FSRS weights for owner", ownerID)
	return nil
}
