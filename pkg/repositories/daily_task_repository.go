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

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 1️⃣ Load existing tasks for this date to preserve their state (e.g. "done")
		var existing []entities.DailyTask
		if err := tx.
			Where("user_id = ? AND task_date = ?", userID, taskDate).
			Find(&existing).Error; err != nil {
			return err
		}

		// Build a map of item_id -> existing state so we don't reset "done" items
		existingStateByItem := make(map[uuid.UUID]string, len(existing))
		for _, e := range existing {
			existingStateByItem[e.ItemID] = e.State
		}

		// Build a set of item_ids already in the snapshot
		existingItemIDs := make(map[uuid.UUID]struct{}, len(existing))
		for _, e := range existing {
			existingItemIDs[e.ItemID] = struct{}{}
		}

		// 2️⃣ Only insert tasks that are NOT already in the snapshot.
		//    This preserves "done" state for items the user already reviewed today.
		var toInsert []entities.DailyTask
		for _, t := range tasks {
			if _, alreadyExists := existingItemIDs[t.ItemID]; alreadyExists {
				// Item already tracked today — keep existing row (and its state)
				continue
			}
			// Preserve state if somehow we have it (shouldn't happen for new items)
			if state, ok := existingStateByItem[t.ItemID]; ok {
				t.State = state
			}
			toInsert = append(toInsert, t)
		}

		if len(toInsert) == 0 {
			return nil
		}

		if err := tx.Create(&toInsert).Error; err != nil {
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
