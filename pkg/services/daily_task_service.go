package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
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
}

func NewDailyTaskService(
	reviewStateRepo repositories.ReviewStateRepository,
	dailyTaskRepo repositories.DailyTaskRepository,
) DailyTaskService {
	return &dailyTaskService{
		reviewStateRepo: reviewStateRepo,
		dailyTaskRepo:   dailyTaskRepo,
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

	// 1️⃣ Ambil kandidat dari Load Control (Core Engine #2)
	candidates, err := s.reviewStateRepo.FindDueByUser(
		ctx,
		userID,
		now,
		limit,
	)
	if err != nil {
		return nil, err
	}

	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}

	// 2️⃣ Bentuk snapshot daily tasks
	tasks := make([]entities.DailyTask, 0, len(candidates))

	for _, c := range candidates {
		tasks = append(tasks, entities.DailyTask{
			ID:       uuid.New(),
			UserID:   c.UserID,
			ItemID:   c.ItemID,
			CardID:   c.CardID,
			TaskDate: taskDate,

			Source: c.Source,
			State:  "pending",

			CreatedAt: now,
		})
	}

	// 3️⃣ Simpan snapshot (IDEMPOTENT)
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

	taskDate := now.Truncate(24 * time.Hour)

	return s.dailyTaskRepo.ListByUserAndDate(
		ctx,
		userID,
		taskDate,
	)
}
