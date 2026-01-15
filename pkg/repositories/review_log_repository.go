package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hifzhun-api/pkg/entities"
)

type ReviewLogRepository interface {
	Create(ctx context.Context, log *entities.ReviewLog) error
	ListByCardID(ctx context.Context, cardID uuid.UUID, limit int) ([]entities.ReviewLog, error)
}

type reviewLogRepository struct {
	db *gorm.DB
}

func NewReviewLogRepository(db *gorm.DB) ReviewLogRepository {
	return &reviewLogRepository{db: db}
}

func (r *reviewLogRepository) Create(
	ctx context.Context,
	log *entities.ReviewLog,
) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *reviewLogRepository) ListByCardID(
	ctx context.Context,
	cardID uuid.UUID,
	limit int,
) ([]entities.ReviewLog, error) {
	var logs []entities.ReviewLog

	q := r.db.WithContext(ctx).
		Where("card_id = ?", cardID).
		Order("reviewed_at desc")

	if limit > 0 {
		q = q.Limit(limit)
	}

	if err := q.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
