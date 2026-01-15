package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"hifzhun-api/pkg/repositories"
)

type LoadControlItem struct {
	ItemID       uuid.UUID
	CardID       uuid.UUID
	State        string
	Stability    float64
	NextReviewAt *time.Time
}

type LoadControlService interface {
	SelectForToday(
		ctx context.Context,
		userID uuid.UUID,
		now time.Time,
		limit int,
	) ([]LoadControlItem, error)
}

type loadControlService struct {
	reviewStateRepo repositories.ReviewStateRepository
}

func NewLoadControlService(
	reviewStateRepo repositories.ReviewStateRepository,
) LoadControlService {
	return &loadControlService{
		reviewStateRepo: reviewStateRepo,
	}
}

func (s *loadControlService) SelectForToday(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
	limit int,
) ([]LoadControlItem, error) {

	states, err := s.reviewStateRepo.FindDueByUser(
		ctx,
		userID,
		now,
		limit,
	)
	if err != nil {
		return nil, err
	}

	items := make([]LoadControlItem, 0, len(states))
	for _, s := range states {
		items = append(items, LoadControlItem{
			ItemID:       s.ItemID,
			CardID:       s.CardID,
			State:        s.State,
			Stability:    s.Stability,
			NextReviewAt: s.NextReviewAt,
		})
	}

	return items, nil
}
