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
	classBookRepo       repositories.ClassBookRepository
	juzItemRepo         *repositories.JuzItemRepository
}

func NewItemReviewService(
	itemRepo *repositories.ItemRepository,
	fsrsWeightsRepo repositories.FSRSWeightsRepository,
	dailyTaskActionRepo repositories.DailyTaskActionRepository,
	classMemberRepo repositories.ClassMemberRepository,
	classRepo repositories.ClassRepository,
	classBookRepo repositories.ClassBookRepository,
	juzItemRepo *repositories.JuzItemRepository,
) *ItemReviewService {
	return &ItemReviewService{
		itemRepo:            itemRepo,
		fsrsWeightsRepo:     fsrsWeightsRepo,
		dailyTaskActionRepo: dailyTaskActionRepo,
		classMemberRepo:     classMemberRepo,
		classRepo:           classRepo,
		classBookRepo:       classBookRepo,
		juzItemRepo:         juzItemRepo,
	}
}

// isItemInActiveQuranClass only returns true for an item created in a specific,
// active Quran class where the owner is still a member.
func (s *ItemReviewService) isItemInActiveQuranClass(item *entities.Item, userID uuid.UUID) bool {
	if item == nil || item.SourceType != "quran" || s.juzItemRepo == nil {
		return false
	}
	infoByItemID, err := s.juzItemRepo.FindJuzInfoByItemIDs([]string{item.ID.String()})
	if err != nil {
		return false
	}
	info, exists := infoByItemID[item.ID.String()]
	if !exists || info.ClassID == nil {
		return false
	}
	class, err := s.classRepo.FindByID(*info.ClassID)
	if err != nil || class.Type != entities.ClassTypeQuran || !class.IsActive {
		return false
	}
	isMember, err := s.classMemberRepo.IsMember(class.ID.String(), userID.String())
	return err == nil && isMember
}

func (s *ItemReviewService) canAccessBookItem(item *entities.Item, userID uuid.UUID) bool {
	if item.SourceType != "book" {
		return true
	}

	bookID, ok := bookIDFromItemContentRef(item.ContentRef)
	if !ok || s.classBookRepo == nil {
		return false
	}

	isClassBook, err := s.classBookRepo.IsBookAssignedToClass(bookID)
	if err != nil {
		return false
	}
	if !isClassBook {
		return true
	}

	isPublished, err := s.classBookRepo.IsBookPublished(bookID)
	if err == nil && isPublished {
		isImported, err := s.classBookRepo.IsBookImportedByUser(bookID, userID.String())
		if err == nil && isImported {
			return true
		}
		isOwner, err := s.classBookRepo.IsBookOwner(bookID, userID.String())
		if err == nil && isOwner {
			return true
		}
	}

	isOwner, err := s.classBookRepo.IsBookOwner(bookID, userID.String())
	if err == nil && isOwner {
		return true
	}

	allowedMember, err := s.classBookRepo.IsBookAccessibleByMember(bookID, userID.String())
	if err == nil && allowedMember {
		return true
	}

	allowedTeacher, err := s.classBookRepo.IsBookAccessibleByTeacher(bookID, userID.String())
	return err == nil && allowedTeacher
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
	if !s.canAccessBookItem(item, userID) {
		return nil, errors.New("you don't have access to this book item")
	}

	// 3. Normalize legacy 'interval' status for quran items to 'fsrs_active'
	if item.SourceType == "quran" && item.Status == entities.ItemStatusInterval {
		item.Status = entities.ItemStatusFSRSActive
		item.IntervalEndAt = &now
		if item.FSRSStartAt == nil {
			if item.IntervalStartAt != nil {
				item.FSRSStartAt = item.IntervalStartAt
			} else {
				item.FSRSStartAt = &now
			}
		}
	}

	// Validate status - must be fsrs_active or graduate (for periodic review)
	// For book items: can also be 'start' or 'menghafal'
	if item.Status != entities.ItemStatusFSRSActive && item.Status != entities.ItemStatusGraduate {
		if item.SourceType != "book" || (item.Status != entities.ItemStatusStart && item.Status != entities.ItemStatusMenghafal) {
			return nil, errors.New("item must be in 'fsrs_active' or 'graduate' status to review")
		}
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

	// 10. Run FSRS review with appropriate retention
	retention := fsrs.DefaultRetention
	if item.SourceType == "quran" {
		retention = fsrs.QuranRetention
		if rating == fsrs.Good && item.Stability <= 30.0 && item.Status != entities.ItemStatusGraduate {
			rating = fsrs.Hard
		}
	}
	result := fsrs.ReviewWithRetention(prevState, rating, now, weights, retention)

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

	// 12. Check for graduation in fsrs_active
	graduated := false
	pendingGraduate := false

	// Handle book items: START → FSRS_ACTIVE
	if item.SourceType == "book" {
		// Transition from start/menghafal to fsrs_active on first review
		if item.Status == entities.ItemStatusStart || item.Status == entities.ItemStatusMenghafal {
			item.Status = entities.ItemStatusFSRSActive
			if item.FSRSStartAt == nil {
				item.FSRSStartAt = &now
			}
		}
	}

	if qualifiesForGraduation(item, now) {
		if item.SourceType == "quran" {
			if s.isItemInActiveQuranClass(item, userID) {
				item.Status = entities.ItemStatusPendingGraduate
				pendingGraduate = true
			} else {
				item.Status = entities.ItemStatusGraduate
				graduated = true
			}
		} else {
			item.Status = entities.ItemStatusGraduate
			graduated = true
		}
	}
	// 13. If item is graduate, handle next review policy
	if item.Status == entities.ItemStatusGraduate {
		if item.SourceType == "quran" {
			// Quran: schedule periodic post-graduation review
			graduateNextReview := now.AddDate(0, 0, entities.GraduateReviewDays)
			graduateNextReview = time.Date(graduateNextReview.Year(), graduateNextReview.Month(), graduateNextReview.Day(), 0, 0, 0, 0, graduateNextReview.Location())
			item.NextReviewAt = &graduateNextReview
			nextReview = graduateNextReview
			intervalDays = entities.GraduateReviewDays
		} else {
			// Book: no further review after graduation
			item.NextReviewAt = nil
		}
	}

	// 14. Save item
	if err := s.itemRepo.Update(item); err != nil {
		return nil, err
	}

	// 15. Mark daily task as done (ignore error if not found)
	taskDate := utils.NormalizeDate(now)
	if s.dailyTaskActionRepo != nil {
		_ = s.dailyTaskActionRepo.UpdateStateByItemID(
			context.Background(),
			userID,
			taskDate,
			itemID,
			"done",
		)
	}

	return &ItemReviewResult{
		Item:            item,
		IntervalDays:    intervalDays,
		NextReviewAt:    &nextReview,
		Graduated:       graduated,
		PendingGraduate: pendingGraduate,
		ReviewCount:     item.ReviewCount,
	}, nil
}
