package services

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
)

type ItemStatusService struct {
	itemRepo *repositories.ItemRepository
}

func NewItemStatusService(itemRepo *repositories.ItemRepository) *ItemStatusService {
	return &ItemStatusService{itemRepo: itemRepo}
}

// StartInterval moves item from menghafal → interval
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

	// Transition to interval
	now := time.Now()
	endAt := now.AddDate(0, 0, intervalDays)

	item.Status = entities.ItemStatusInterval
	item.IntervalDays = intervalDays
	item.IntervalStartAt = &now
	item.IntervalEndAt = &endAt

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
