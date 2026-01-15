package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hifzhun-api/pkg/entities"
)

type DailyTaskRepository interface {
	UpsertDailyTasks(
		ctx context.Context,
		userID uuid.UUID,
		taskDate time.Time,
		tasks []entities.DailyTask,
	) error

	ListByUserAndDate(
		ctx context.Context,
		userID uuid.UUID,
		taskDate time.Time,
	) ([]entities.DailyTask, error)
}

type dailyTaskRepository struct {
	db *gorm.DB
}

func NewDailyTaskRepository(db *gorm.DB) DailyTaskRepository {
	return &dailyTaskRepository{db: db}
}

func (r *dailyTaskRepository) UpsertDailyTasks(
	ctx context.Context,
	userID uuid.UUID,
	taskDate time.Time,
	tasks []entities.DailyTask,
) error {

	if len(tasks) == 0 {
		return nil
	}

	// Gunakan transaction agar konsisten
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 1️⃣ Hapus snapshot lama (AMAN karena 1 hari)
		if err := tx.
			Where("user_id = ?", userID).
			Where("task_date = ?", taskDate).
			Delete(&entities.DailyTask{}).Error; err != nil {
			return err
		}

		// 2️⃣ Insert snapshot baru
		if err := tx.Create(&tasks).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *dailyTaskRepository) ListByUserAndDate(
	ctx context.Context,
	userID uuid.UUID,
	taskDate time.Time,
) ([]entities.DailyTask, error) {

	var tasks []entities.DailyTask

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("task_date = ?", taskDate).
		Order("created_at ASC").
		Find(&tasks).Error

	return tasks, err
}
