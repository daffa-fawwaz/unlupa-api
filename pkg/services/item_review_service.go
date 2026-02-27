package services

import (
	"context"
	"errors"
	"fmt"
	"math"
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
	PendingGraduate bool // true if waiting for teacher approval
	ReviewCount     int  // total reviews for this item
}

type ItemReviewService struct {
	itemRepo            *repositories.ItemRepository
	fsrsWeightsRepo     repositories.FSRSWeightsRepository
	dailyTaskActionRepo repositories.DailyTaskActionRepository
	classMemberRepo     repositories.ClassMemberRepository
	classRepo           repositories.ClassRepository
}

func NewItemReviewService(
	itemRepo *repositories.ItemRepository,
	fsrsWeightsRepo repositories.FSRSWeightsRepository,
	dailyTaskActionRepo repositories.DailyTaskActionRepository,
	classMemberRepo repositories.ClassMemberRepository,
	classRepo repositories.ClassRepository,
) *ItemReviewService {
	return &ItemReviewService{
		itemRepo:            itemRepo,
		fsrsWeightsRepo:     fsrsWeightsRepo,
		dailyTaskActionRepo: dailyTaskActionRepo,
		classMemberRepo:     classMemberRepo,
		classRepo:           classRepo,
	}
}

// isUserInQuranClass checks if user has joined any active Quran class
func (s *ItemReviewService) isUserInQuranClass(userID uuid.UUID) bool {
	classes, err := s.classMemberRepo.FindByUserID(userID.String())
	if err != nil || len(classes) == 0 {
		return false
	}

	// Check if any of the classes is a Quran-type class
	for _, membership := range classes {
		class, err := s.classRepo.FindByID(membership.ClassID.String())
		if err != nil {
			continue
		}
		if class.Type == entities.ClassTypeQuran && class.IsActive {
			return true
		}
	}
	return false
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

	// 3. Validate status - must be fsrs_active or graduate (for periodic review)
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
	if !isFirstReview && item.NextReviewAt != nil {
		if now.Before(*item.NextReviewAt) {
			return nil, fmt.Errorf("review not allowed yet, next review at: %s", item.NextReviewAt.Format("2006-01-02 15:04"))
		}
	}

	// 7. Set initial FSRS state if first review
	if isFirstReview {
		item.Stability = 0.4
		item.Difficulty = 5.0
	}
	// Guard for items coming from interval phase: they can have LastReviewAt set
	// but still have zero/invalid FSRS params, which can produce NaN.
	if math.IsNaN(item.Stability) || math.IsInf(item.Stability, 0) || item.Stability <= 0 {
		item.Stability = 0.4
	}
	if math.IsNaN(item.Difficulty) || math.IsInf(item.Difficulty, 0) || item.Difficulty <= 0 {
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
	if math.IsNaN(item.Stability) || math.IsInf(item.Stability, 0) || item.Stability <= 0 {
		item.Stability = 0.01
	}
	if math.IsNaN(item.Difficulty) || math.IsInf(item.Difficulty, 0) || item.Difficulty <= 0 {
		item.Difficulty = 5.0
	}
	item.ReviewCount++
	item.LastReviewAt = &now

	intervalDays := int(result.Interval.Hours() / 24)
	// Normalize next review time to 00:00:00
	nextReview := now.Add(result.Interval)
	nextReview = time.Date(nextReview.Year(), nextReview.Month(), nextReview.Day(), 0, 0, 0, 0, nextReview.Location())
	item.NextReviewAt = &nextReview

	// 12. Check for graduation - ONLY for quran items in fsrs_active
	// Book items stay in fsrs_active forever (no auto-graduation)
	// Graduation is TIME-BASED: item must have been in fsrs_active for >= 30 days
	graduated := false
	pendingGraduate := false
	if item.Status == entities.ItemStatusFSRSActive &&
		item.SourceType == "quran" {

		// Calculate how many days the item has been in fsrs_active phase
		// FSRSStartAt = when item entered fsrs_active (day 1)
		daysInFSRSActive := 0
		if item.FSRSStartAt != nil {
			daysInFSRSActive = int(now.Sub(*item.FSRSStartAt).Hours() / 24)
		} else if item.IntervalEndAt != nil {
			// Fallback for legacy records
			daysInFSRSActive = int(now.Sub(*item.IntervalEndAt).Hours() / 24)
		}

		// Graduate if days in fsrs_active >= threshold OR stability >= threshold
		if daysInFSRSActive >= entities.GraduationIntervalDays || item.Stability >= entities.GraduateStabilityThreshold {
			// Check if user is in a Quran class - if yes, require teacher approval
			if s.isUserInQuranClass(userID) {
				item.Status = entities.ItemStatusPendingGraduate
				pendingGraduate = true
			} else {
				item.Status = entities.ItemStatusGraduate
				graduated = true
			}
		}
	}

	// 13. If item is graduate (just graduated or already graduate), set next review to 20 days
	if item.Status == entities.ItemStatusGraduate {
		graduateNextReview := now.AddDate(0, 0, entities.GraduateReviewDays)
		// Normalize next review time to 00:00:00
		graduateNextReview = time.Date(graduateNextReview.Year(), graduateNextReview.Month(), graduateNextReview.Day(), 0, 0, 0, 0, graduateNextReview.Location())
		item.NextReviewAt = &graduateNextReview
		nextReview = graduateNextReview
		intervalDays = entities.GraduateReviewDays
	}

	// 14. Save item
	if err := s.itemRepo.Update(item); err != nil {
		return nil, err
	}

	// 15. Mark daily task as done (ignore error if not found)
	taskDate := utils.NormalizeDate(now)
	_ = s.dailyTaskActionRepo.UpdateStateByItemID(
		context.Background(),
		userID,
		taskDate,
		itemID,
		"done",
	)

	return &ItemReviewResult{
		Item:            item,
		IntervalDays:    intervalDays,
		NextReviewAt:    &nextReview,
		Graduated:       graduated,
		PendingGraduate: pendingGraduate,
		ReviewCount:     item.ReviewCount,
	}, nil
}
