package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hifzhun-api/pkg/entities"
)

type ReviewStateRepository interface {
	FindDueByUser(
		ctx context.Context,
		userID uuid.UUID,
		now time.Time,
		limit int,
	) ([]entities.ReviewState, error)

	Upsert(
		ctx context.Context,
		state *entities.ReviewState,
	) error

	FindByCardID(ctx context.Context, cardID uuid.UUID) (*entities.ReviewState, error)

}

type reviewStateRepository struct {
	db *gorm.DB
}

func NewReviewStateRepository(db *gorm.DB) ReviewStateRepository {
	return &reviewStateRepository{db: db}
}

func (r *reviewStateRepository) FindDueByUser(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
	limit int,
) ([]entities.ReviewState, error) {

	var states []entities.ReviewState

	q := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("state = ?", "active").
		Where("next_review_at IS NOT NULL").
		Where("next_review_at <= ?", now).
		Order("next_review_at ASC, stability ASC")

	if limit > 0 {
		q = q.Limit(limit)
	}

	if err := q.Find(&states).Error; err != nil {
		return nil, err
	}

	return states, nil
}

func (r *reviewStateRepository) Upsert(
	ctx context.Context,
	state *entities.ReviewState,
) error {
	return r.db.WithContext(ctx).
		Save(state).
		Error
}

func (r *reviewStateRepository) FindByCardID(
	ctx context.Context,
	cardID uuid.UUID,
) (*entities.ReviewState, error) {

	var rs entities.ReviewState
	err := r.db.WithContext(ctx).
		Where("card_id = ?", cardID).
		First(&rs).Error

	if err != nil {
		return nil, err
	}
	return &rs, nil
}

