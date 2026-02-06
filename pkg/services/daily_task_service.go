package services

import (
	"context"
	"time"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/utils"

	"github.com/google/uuid"
)

type DailyTaskService interface {
	GenerateToday(
		ctx context.Context,
		userID uuid.UUID,
		now time.Time,
		limit int,
	) ([]entities.DailyTask, error)

	ListToday(
		ctx context.Context,
		userID uuid.UUID,
		now time.Time,
	) ([]entities.DailyTask, error)
}

type dailyTaskService struct {
	reviewStateRepo repositories.ReviewStateRepository
	dailyTaskRepo   repositories.DailyTaskRepository
	itemRepo        *repositories.ItemRepository
}

func NewDailyTaskService(
	reviewStateRepo repositories.ReviewStateRepository,
	dailyTaskRepo repositories.DailyTaskRepository,
	itemRepo *repositories.ItemRepository,
) DailyTaskService {
	return &dailyTaskService{
		reviewStateRepo: reviewStateRepo,
		dailyTaskRepo:   dailyTaskRepo,
		itemRepo:        itemRepo,
	}
}

func (s *dailyTaskService) GenerateToday(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
	limit int,
) ([]entities.DailyTask, error) {

	// 📌 SNAPSHOT DATE (WAJIB)
	taskDate := now.Truncate(24 * time.Hour)

	tasks := make([]entities.DailyTask, 0)

	// ========== 1️⃣ Items dari Interval yang sudah deadline ==========
	intervalItems, err := s.itemRepo.FindIntervalDeadlineReached(now)
	if err != nil {
		return nil, err
	}

	for _, item := range intervalItems {
		// Filter by owner
		if item.OwnerID != userID {
			continue
		}

		tasks = append(tasks, entities.DailyTask{
			ID:       uuid.New(),
			UserID:   userID,
			ItemID:   item.ID,
			CardID:   uuid.Nil, // No card yet for interval items
			TaskDate: taskDate,
			Source:   "interval", // Mark as interval source
			State:    "pending",
			CreatedAt: now,
		})

		// Update item status to fsrs_active
		item.Status = entities.ItemStatusFSRSActive
		s.itemRepo.Update(&item)
	}

	// ========== 2️⃣ Items FSRS Active yang due untuk review ==========
	fsrsItems, err := s.itemRepo.FindFSRSDueItems(userID, now)
	if err != nil {
		return nil, err
	}

	for _, item := range fsrsItems {
		tasks = append(tasks, entities.DailyTask{
			ID:       uuid.New(),
			UserID:   userID,
			ItemID:   item.ID,
			CardID:   uuid.Nil, // Item-based, no card
			TaskDate: taskDate,
			Source:   "quran", // or item.SourceType
			State:    "pending",
			CreatedAt: now,
		})
	}

	// ========== 3️⃣ Cards dari FSRS Review State (existing logic) ==========
	candidates, err := s.reviewStateRepo.FindDueByUser(
		ctx,
		userID,
		now,
		limit,
	)
	if err != nil {
		return nil, err
	}

	for _, c := range candidates {
		tasks = append(tasks, entities.DailyTask{
			ID:       uuid.New(),
			UserID:   c.UserID,
			ItemID:   c.ItemID,
			CardID:   c.CardID,
			TaskDate: taskDate,
			Source:   c.Source,
			State:    "pending",
			CreatedAt: now,
		})
	}

	// ========== 4️⃣ Graduate items untuk review bulanan (by juz) ==========
	// Juz 1 → tanggal 1, Juz 2 → tanggal 2, dst
	dayOfMonth := now.Day()
	if dayOfMonth <= 30 { // Only juz 1-30 exist
		gradItems, err := s.itemRepo.FindGraduateItemsByJuzDay(userID, dayOfMonth)
		if err != nil {
			return nil, err
		}

		for _, item := range gradItems {
			tasks = append(tasks, entities.DailyTask{
				ID:        uuid.New(),
				UserID:    userID,
				ItemID:    item.ID,
				CardID:    uuid.Nil,
				TaskDate:  taskDate,
				Source:    "graduate", // Mark as graduate review
				State:     "pending",
				CreatedAt: now,
			})
		}
	}

	// Apply limit if needed
	if limit > 0 && len(tasks) > limit {
		tasks = tasks[:limit]
	}

	// 5️⃣ Simpan snapshot (IDEMPOTENT)
	if err := s.dailyTaskRepo.UpsertDailyTasks(
		ctx,
		userID,
		taskDate,
		tasks,
	); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *dailyTaskService) ListToday(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
) ([]entities.DailyTask, error) {

	taskDate := utils.NormalizeDate(now)

	return s.dailyTaskRepo.ListByUserAndDate(
		ctx,
		userID,
		taskDate,
	)
}

