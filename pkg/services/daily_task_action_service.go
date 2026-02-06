package services

import (
	"context"
	"time"

	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/utils"

	"github.com/google/uuid"
)

type DailyTaskActionService interface {
	MarkDone(
		ctx context.Context,
		userID uuid.UUID,
		cardID uuid.UUID,
		now time.Time,
	) error

	MarkSkipped(
		ctx context.Context,
		userID uuid.UUID,
		cardID uuid.UUID,
		now time.Time,
	) error

	MarkDoneByItemID(
		ctx context.Context,
		userID uuid.UUID,
		itemID uuid.UUID,
		now time.Time,
	) error
}

type dailyTaskActionService struct {
	repo repositories.DailyTaskActionRepository
}

func NewDailyTaskActionService(
	repo repositories.DailyTaskActionRepository,
) DailyTaskActionService {
	return &dailyTaskActionService{repo: repo}
}

func (s *dailyTaskActionService) MarkDone(
	ctx context.Context,
	userID uuid.UUID,
	cardID uuid.UUID,
	now time.Time,
) error {

	taskDate := utils.NormalizeDate(now)

	return s.repo.UpdateState(
		ctx,
		userID,
		taskDate,
		cardID,
		"done",
	)
}

func (s *dailyTaskActionService) MarkSkipped(
	ctx context.Context,
	userID uuid.UUID,
	cardID uuid.UUID,
	now time.Time,
) error {

	taskDate := utils.NormalizeDate(now)

	return s.repo.UpdateState(
		ctx,
		userID,
		taskDate,
		cardID,
		"skipped",
	)
}

// MarkDoneByItemID marks daily task as done by itemID (for interval items without cardID)
func (s *dailyTaskActionService) MarkDoneByItemID(
	ctx context.Context,
	userID uuid.UUID,
	itemID uuid.UUID,
	now time.Time,
) error {

	taskDate := utils.NormalizeDate(now)

	return s.repo.UpdateStateByItemID(
		ctx,
		userID,
		taskDate,
		itemID,
		"done",
	)
}


