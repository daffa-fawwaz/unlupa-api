package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hifzhun-api/pkg/entities"
)

type DailyTaskActionRepository interface {
	UpdateState(
		ctx context.Context,
		userID uuid.UUID,
		taskDate time.Time,
		cardID uuid.UUID,
		newState string,
	) error
}

type dailyTaskActionRepository struct {
	db *gorm.DB
}

func NewDailyTaskActionRepository(db *gorm.DB) DailyTaskActionRepository {
	return &dailyTaskActionRepository{db: db}
}

func (r *dailyTaskActionRepository) UpdateState(
	ctx context.Context,
	userID uuid.UUID,
	taskDate time.Time,
	cardID uuid.UUID,
	newState string,
) error {

	res := r.db.WithContext(ctx).
		Model(&entities.DailyTask{}).
		Where("user_id = ?", userID).
		Where("task_date = ?", taskDate).
		Where("card_id = ?", cardID).
		Where("state = ?", "pending").
		Update("state", newState)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
