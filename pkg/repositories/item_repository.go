package repositories

import (
	"time"

	"hifzhun-api/pkg/entities"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

type ItemRepository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) *ItemRepository {
	return &ItemRepository{db}
}

func (r *ItemRepository) Create(item *entities.Item) error {
	return r.db.Create(item).Error
}

func (r *ItemRepository) GetByID(id uuid.UUID) (*entities.Item, error) {
	var item entities.Item
	err := r.db.Where("id = ?", id).First(&item).Error
	return &item, err
}

func (r *ItemRepository) FindByOwnerAndStatus(ownerID uuid.UUID, status string) ([]entities.Item, error) {
	var items []entities.Item
	err := r.db.Where("owner_id = ? AND status = ?", ownerID, status).Find(&items).Error
	return items, err
}

func (r *ItemRepository) Update(item *entities.Item) error {
	return r.db.Save(item).Error
}

// FindIntervalDeadlineReached finds items with status=interval and deadline reached
func (r *ItemRepository) FindIntervalDeadlineReached(now time.Time) ([]entities.Item, error) {
	var items []entities.Item
	err := r.db.
		Where("status = ? AND interval_end_at <= ?", entities.ItemStatusInterval, now).
		Find(&items).Error
	return items, err
}

// FindIntervalReviewDue finds items with status=interval and interval_next_review_at <= now
func (r *ItemRepository) FindIntervalReviewDue(ownerID uuid.UUID, now time.Time) ([]entities.Item, error) {
	var items []entities.Item
	err := r.db.
		Where("owner_id = ? AND status = ? AND interval_next_review_at <= ?", ownerID, entities.ItemStatusInterval, now).
		Find(&items).Error
	return items, err
}

// FindFSRSDueItems finds items with status=fsrs_active and next_review_at <= now
func (r *ItemRepository) FindFSRSDueItems(ownerID uuid.UUID, now time.Time) ([]entities.Item, error) {
	var items []entities.Item
	err := r.db.
		Where("owner_id = ? AND status = ? AND next_review_at <= ?", ownerID, entities.ItemStatusFSRSActive, now).
		Find(&items).Error
	return items, err
}

// FindGraduateItemsByJuzDay finds graduated items where juz.index = dayOfMonth
func (r *ItemRepository) FindGraduateItemsByJuzDay(ownerID uuid.UUID, dayOfMonth int) ([]entities.Item, error) {
	var items []entities.Item
	err := r.db.
		Joins("JOIN juz_items ON juz_items.item_id = items.id").
		Joins("JOIN juzs ON juzs.id = juz_items.juz_id").
		Where("items.owner_id = ? AND items.status = ? AND juzs.index = ?",
			ownerID, entities.ItemStatusGraduate, dayOfMonth).
		Find(&items).Error
	return items, err
}

// FindByOwner finds all items belonging to a user
func (r *ItemRepository) FindByOwner(ownerID string) ([]entities.Item, error) {
	var items []entities.Item
	err := r.db.Where("owner_id = ?", ownerID).Find(&items).Error
	return items, err
}

// FindByOwnerAndContentRef finds items by owner and content_ref (for duplicate check)
func (r *ItemRepository) FindByOwnerAndContentRef(ownerID uuid.UUID, contentRef string) ([]entities.Item, error) {
	var items []entities.Item
	err := r.db.Where("owner_id = ? AND content_ref = ?", ownerID, contentRef).Find(&items).Error
	return items, err
}

// FindByIDs finds items by a list of IDs
func (r *ItemRepository) FindByIDs(ids []uuid.UUID) ([]entities.Item, error) {
	var items []entities.Item
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

// FindByOwnerAndSourceType finds items by owner and source_type
// FindEligibleForGraduation finds items in fsrs_active that have been there for >= thresholdDays
func (r *ItemRepository) FindEligibleForGraduation(ownerID uuid.UUID, thresholdDays int, now time.Time) ([]entities.Item, error) {
	cutoff := now.AddDate(0, 0, -thresholdDays)
	var items []entities.Item
	// Use fsrs_start_at as the start of fsrs_active phase (day 1).
	// Fallback to interval_end_at for backward compatibility if fsrs_start_at is NULL.
	err := r.db.
		Where("owner_id = ? AND status = ? AND source_type = 'quran'", ownerID, entities.ItemStatusFSRSActive).
		Where("(fsrs_start_at IS NOT NULL AND fsrs_start_at <= ?) OR (fsrs_start_at IS NULL AND interval_end_at IS NOT NULL AND interval_end_at <= ?)", cutoff, cutoff).
		Find(&items).Error
	return items, err
}

func (r *ItemRepository) FindByOwnerAndSourceType(ownerID uuid.UUID, sourceType string) ([]entities.Item, error) {
	var items []entities.Item
	err := r.db.Where("owner_id = ? AND source_type = ?", ownerID, sourceType).
		Order("created_at ASC").
		Find(&items).Error
	return items, err
}

// FindEligibleForGraduationByStability finds fsrs_active Quran items with stability >= threshold
func (r *ItemRepository) FindEligibleForGraduationByStability(ownerID uuid.UUID, stabilityThreshold float64) ([]entities.Item, error) {
	var items []entities.Item
	err := r.db.
		Where("owner_id = ? AND status = ? AND source_type = 'quran' AND stability >= ?",
			ownerID, entities.ItemStatusFSRSActive, stabilityThreshold).
		Find(&items).Error
	return items, err
}
