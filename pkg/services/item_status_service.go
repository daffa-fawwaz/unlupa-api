package services

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
)

type ItemStatusService struct {
	itemRepo            *repositories.ItemRepository
	intervalReviewRepo  *repositories.IntervalReviewLogRepository
}

func NewItemStatusService(
	itemRepo *repositories.ItemRepository,
	intervalReviewRepo *repositories.IntervalReviewLogRepository,
) *ItemStatusService {
	return &ItemStatusService{
		itemRepo:           itemRepo,
		intervalReviewRepo: intervalReviewRepo,
	}
}

// StartInterval moves item from menghafal → interval (recurring review)
func (s *ItemStatusService) StartInterval(itemID uuid.UUID, userID uuid.UUID, intervalDays int) (*entities.Item, error) {
	item, err := s.itemRepo.GetByID(itemID)
	if err != nil {
		return nil, errors.New("item not found")
	}

	// Validate ownership
	if item.OwnerID != userID {
		return nil, errors.New("unauthorized")
	}

	// Validate current status
	if item.Status != entities.ItemStatusMenghafal {
		return nil, errors.New("item must be in 'menghafal' status to start interval")
	}

	// Validate interval days
	if intervalDays < 1 {
		return nil, errors.New("interval_days must be at least 1")
	}

	// Transition to interval with recurring review
	now := time.Now()

// Target date (hari + interval)
targetDate := now.AddDate(0, 0, intervalDays)

// Normalize ke jam 00:00:00
nextReview := time.Date(
	targetDate.Year(),
	targetDate.Month(),
	targetDate.Day(),
	0, 0, 0, 0,
	targetDate.Location(),
)

	item.Status = entities.ItemStatusInterval
	item.IntervalDays = intervalDays
	item.IntervalStartAt = &now
	item.IntervalNextReviewAt = &nextReview

	if err := s.itemRepo.Update(item); err != nil {
		return nil, err
	}

	return item, nil
}

// IntervalReviewResult represents the result of an interval review
type IntervalReviewResult struct {
	Item         *entities.Item `json:"item"`
	NextReviewAt *time.Time     `json:"next_review_at"`
	Rating       int            `json:"rating"`
}

// ReviewInterval reviews an item in the interval phase
func (s *ItemStatusService) ReviewInterval(itemID uuid.UUID, userID uuid.UUID, rating int) (*IntervalReviewResult, error) {
	item, err := s.itemRepo.GetByID(itemID)
	if err != nil {
		return nil, errors.New("item not found")
	}

	// Validate ownership
	if item.OwnerID != userID {
		return nil, errors.New("unauthorized")
	}

	// Validate status
	if item.Status != entities.ItemStatusInterval {
		return nil, errors.New("item must be in 'interval' status to review")
	}

	// Validate rating (1=bad, 2=good, 3=perfect)
	if rating < 1 || rating > 3 {
		return nil, errors.New("rating must be between 1 and 3 (1=bad, 2=good, 3=perfect)")
	}

	// Check if review is allowed (now >= interval_next_review_at)
	now := time.Now()
	if item.IntervalNextReviewAt != nil && now.Before(*item.IntervalNextReviewAt) {
		return nil, errors.New("review not allowed yet, next review at: " + item.IntervalNextReviewAt.Format("2006-01-02 15:04"))
	}

	// Save review log
	reviewLog := &entities.IntervalReviewLog{
		UserID:     userID,
		ItemID:     itemID,
		Rating:     rating,
		ReviewedAt: now,
	}
	if err := s.intervalReviewRepo.Create(reviewLog); err != nil {
		return nil, err
	}

	// Update next review date

	targetDate := now.AddDate(0, 0, item.IntervalDays)

	nextReview := time.Date(
		targetDate.Year(),
		targetDate.Month(),
		targetDate.Day(),
		0, 0, 0, 0,
		targetDate.Location(),
	)
	item.IntervalNextReviewAt = &nextReview
	item.ReviewCount++
	item.LastReviewAt = &now

	if err := s.itemRepo.Update(item); err != nil {
		return nil, err
	}

	return &IntervalReviewResult{
		Item:         item,
		NextReviewAt: &nextReview,
		Rating:       rating,
	}, nil
}

// IntervalStatsResult represents interval review statistics
type IntervalStatsResult struct {
	AverageRating float64 `json:"average_rating"`
	TotalReviews  int     `json:"total_reviews"`
	Performance   string  `json:"performance"` // bad, good, perfect
}

// GetIntervalReviewStats calculates the average rating and performance label
func (s *ItemStatusService) GetIntervalReviewStats(itemID uuid.UUID, userID uuid.UUID) (*IntervalStatsResult, error) {
	item, err := s.itemRepo.GetByID(itemID)
	if err != nil {
		return nil, errors.New("item not found")
	}

	// Validate ownership
	if item.OwnerID != userID {
		return nil, errors.New("unauthorized")
	}

	avg, count, err := s.intervalReviewRepo.GetAverageRatingByItemID(itemID)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return &IntervalStatsResult{
			AverageRating: 0,
			TotalReviews:  0,
			Performance:   "no_reviews",
		}, nil
	}

	// Determine performance label
	var performance string
	if avg < 1.5 {
		performance = "bad"
	} else if avg < 2.5 {
		performance = "good"
	} else {
		performance = "perfect"
	}

	return &IntervalStatsResult{
		AverageRating: avg,
		TotalReviews:  count,
		Performance:   performance,
	}, nil
}

// ActivateToFSRS moves item from interval → fsrs_active (user decision)
func (s *ItemStatusService) ActivateToFSRS(itemID uuid.UUID, userID uuid.UUID) (*entities.Item, error) {
	item, err := s.itemRepo.GetByID(itemID)
	if err != nil {
		return nil, errors.New("item not found")
	}

	if item.OwnerID != userID {
		return nil, errors.New("unauthorized")
	}

	if item.Status != entities.ItemStatusInterval {
		return nil, errors.New("item must be in 'interval' status to activate FSRS")
	}

	now := time.Now()

	// Hitung next review berdasarkan interval sebelumnya
	targetDate := now.AddDate(0, 0, item.IntervalDays)

	// Normalize ke 00:00
	nextReview := time.Date(
		targetDate.Year(),
		targetDate.Month(),
		targetDate.Day(),
		0, 0, 0, 0,
		targetDate.Location(),
	)

	item.Status = entities.ItemStatusFSRSActive
	item.IntervalEndAt = &now
	item.NextReviewAt = &nextReview // ✅ INI YANG PENTING

	if err := s.itemRepo.Update(item); err != nil {
		return nil, err
	}

	return item, nil
}

// GetItemsByStatus returns items by status for a user
func (s *ItemStatusService) GetItemsByStatus(userID uuid.UUID, status string) ([]entities.Item, error) {
	// Validate status
	validStatuses := map[string]bool{
		entities.ItemStatusMenghafal:       true,
		entities.ItemStatusInterval:        true,
		entities.ItemStatusFSRSActive:      true,
		entities.ItemStatusPendingGraduate: true,
		entities.ItemStatusGraduate:        true,
		entities.ItemStatusInactive:        true,
	}

	if !validStatuses[status] {
		return nil, errors.New("invalid status")
	}

	return s.itemRepo.FindByOwnerAndStatus(userID, status)
}

// ActivateFSRS moves item from interval → fsrs_active (called by scheduler)
func (s *ItemStatusService) ActivateFSRS(item *entities.Item) error {
	if item.Status != entities.ItemStatusInterval {
		return errors.New("item must be in 'interval' status")
	}

	item.Status = entities.ItemStatusFSRSActive
	return s.itemRepo.Update(item)
}

// Graduate moves item to graduate status
func (s *ItemStatusService) Graduate(item *entities.Item) error {
	if item.Status != entities.ItemStatusFSRSActive {
		return errors.New("item must be in 'fsrs_active' status")
	}

	item.Status = entities.ItemStatusGraduate
	return s.itemRepo.Update(item)
}

// GetDeadlineItems returns items that have reached their interval deadline (view only)
func (s *ItemStatusService) GetDeadlineItems(userID uuid.UUID) ([]entities.Item, error) {
	now := time.Now()
	allItems, err := s.itemRepo.FindIntervalDeadlineReached(now)
	if err != nil {
		return nil, err
	}

	// Filter by user
	userItems := make([]entities.Item, 0)
	for _, item := range allItems {
		if item.OwnerID == userID {
			userItems = append(userItems, item)
		}
	}

	return userItems, nil
}

// DeactivateItem moves book item from fsrs_active → inactive
// Only for non-quran items (book items)
func (s *ItemStatusService) DeactivateItem(itemID uuid.UUID, userID uuid.UUID) (*entities.Item, error) {
	item, err := s.itemRepo.GetByID(itemID)
	if err != nil {
		return nil, errors.New("item not found")
	}

	// Validate ownership
	if item.OwnerID != userID {
		return nil, errors.New("unauthorized")
	}

	// Validate source type - only for book items
	if item.SourceType == "quran" {
		return nil, errors.New("quran items cannot be deactivated")
	}

	// Validate current status - must be fsrs_active
	if item.Status != entities.ItemStatusFSRSActive {
		return nil, errors.New("item must be in 'fsrs_active' status to deactivate")
	}

	// Transition to inactive
	item.Status = entities.ItemStatusInactive

	if err := s.itemRepo.Update(item); err != nil {
		return nil, err
	}

	return item, nil
}

// ReactivateItem moves book item from inactive → fsrs_active
// Only for non-quran items (book items)
func (s *ItemStatusService) ReactivateItem(itemID uuid.UUID, userID uuid.UUID) (*entities.Item, error) {
	item, err := s.itemRepo.GetByID(itemID)
	if err != nil {
		return nil, errors.New("item not found")
	}

	// Validate ownership
	if item.OwnerID != userID {
		return nil, errors.New("unauthorized")
	}

	// Validate source type - only for book items
	if item.SourceType == "quran" {
		return nil, errors.New("quran items cannot be reactivated")
	}

	// Validate current status - must be inactive
	if item.Status != entities.ItemStatusInactive {
		return nil, errors.New("item must be in 'inactive' status to reactivate")
	}

	// Transition back to fsrs_active
	item.Status = entities.ItemStatusFSRSActive

	if err := s.itemRepo.Update(item); err != nil {
		return nil, err
	}

	return item, nil
}
