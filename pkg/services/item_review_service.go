package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/fsrs"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/utils"
)

type ItemReviewResult struct {
	Item            *entities.Item
	IntervalDays    int
	NextReviewAt    *time.Time
	Graduated       bool
}

type ItemReviewService struct {
	itemRepo            *repositories.ItemRepository
	fsrsWeightsRepo     repositories.FSRSWeightsRepository
	dailyTaskActionRepo repositories.DailyTaskActionRepository
}

func NewItemReviewService(
	itemRepo *repositories.ItemRepository,
	fsrsWeightsRepo repositories.FSRSWeightsRepository,
	dailyTaskActionRepo repositories.DailyTaskActionRepository,
) *ItemReviewService {
	return &ItemReviewService{
		itemRepo:            itemRepo,
		fsrsWeightsRepo:     fsrsWeightsRepo,
		dailyTaskActionRepo: dailyTaskActionRepo,
	}
}

func (s *ItemReviewService) ReviewItem(
	userID uuid.UUID,
	itemID uuid.UUID,
	rating fsrs.Rating,
	now time.Time,
) (*ItemReviewResult, error) {

	// 1. Get item
	item, err := s.itemRepo.GetByID(itemID)
	if err != nil {
		return nil, errors.New("item not found")
	}

	// 2. Validate ownership
	if item.OwnerID != userID {
		return nil, errors.New("unauthorized")
	}

	// 3. Validate status - must be fsrs_active or graduate (for monthly review)
	if item.Status != entities.ItemStatusFSRSActive && item.Status != entities.ItemStatusGraduate {
		return nil, errors.New("item must be in 'fsrs_active' or 'graduate' status to review")
	}

	// 4. Validate rating
	if rating < fsrs.Again || rating > fsrs.Easy {
		return nil, errors.New("invalid rating (1-4)")
	}

	// 5. Check if first review
	isFirstReview := item.LastReviewAt == nil

	// 6. Check if review is allowed (must be first review OR now >= next_review_at)
	// Skip this check for graduate items (monthly review)
	if item.Status == entities.ItemStatusFSRSActive && !isFirstReview && item.NextReviewAt != nil {
		if now.Before(*item.NextReviewAt) {
			return nil, fmt.Errorf("review not allowed yet, next review at: %s", item.NextReviewAt.Format("2006-01-02 15:04"))
		}
	}

	// 7. Set initial FSRS state if first review
	if isFirstReview {
		item.Stability = 0.4
		item.Difficulty = 5.0
	}

	// 8. Use default FSRS weights (simpler, no DB query needed)
	weights := fsrs.DefaultWeights()

	// 9. Prepare previous state
	var lastReview time.Time
	if item.LastReviewAt != nil {
		lastReview = *item.LastReviewAt
	}

	prevState := fsrs.CardState{
		Stability:  item.Stability,
		Difficulty: item.Difficulty,
		LastReview: lastReview,
	}

	// 10. Run FSRS review
	result := fsrs.Review(prevState, rating, now, weights)

	// 11. Update item with new FSRS state
	item.Stability = result.NewState.Stability
	item.Difficulty = result.NewState.Difficulty
	item.LastReviewAt = &now

	intervalDays := int(result.Interval.Hours() / 24)
	nextReview := now.Add(result.Interval)
	item.NextReviewAt = &nextReview

	// 12. Check for graduation (interval >= 30 days) - only for fsrs_active
	graduated := false
	if item.Status == entities.ItemStatusFSRSActive && intervalDays >= entities.GraduationIntervalDays {
		item.Status = entities.ItemStatusGraduate
		graduated = true
	}

	// 13. Save item
	if err := s.itemRepo.Update(item); err != nil {
		return nil, err
	}

	// 14. Mark daily task as done (ignore error if not found)
	taskDate := utils.NormalizeDate(now)
	_ = s.dailyTaskActionRepo.UpdateStateByItemID(
		context.Background(),
		userID,
		taskDate,
		itemID,
		"done",
	)

	return &ItemReviewResult{
		Item:         item,
		IntervalDays: intervalDays,
		NextReviewAt: &nextReview,
		Graduated:    graduated,
	}, nil
}


