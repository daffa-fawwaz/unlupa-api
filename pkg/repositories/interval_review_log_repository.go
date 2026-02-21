package repositories

import (
	"hifzhun-api/pkg/entities"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IntervalReviewLogRepository struct {
	db *gorm.DB
}

func NewIntervalReviewLogRepository(db *gorm.DB) *IntervalReviewLogRepository {
	return &IntervalReviewLogRepository{db: db}
}

func (r *IntervalReviewLogRepository) Create(log *entities.IntervalReviewLog) error {
	return r.db.Create(log).Error
}

func (r *IntervalReviewLogRepository) FindByItemID(itemID uuid.UUID) ([]entities.IntervalReviewLog, error) {
	var logs []entities.IntervalReviewLog
	err := r.db.Where("item_id = ?", itemID).Order("reviewed_at DESC").Find(&logs).Error
	return logs, err
}

// GetAverageRatingByItemID returns average rating and total count for an item
func (r *IntervalReviewLogRepository) GetAverageRatingByItemID(itemID uuid.UUID) (float64, int, error) {
	var result struct {
		Avg   float64
		Count int
	}

	err := r.db.Model(&entities.IntervalReviewLog{}).
		Select("COALESCE(AVG(rating), 0) as avg, COUNT(*) as count").
		Where("item_id = ?", itemID).
		Scan(&result).Error

	return result.Avg, result.Count, err
}
