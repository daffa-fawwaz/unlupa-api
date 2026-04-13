package repositories

import (
	"hifzhun-api/pkg/entities"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookUpdateRequestRepository struct {
	db *gorm.DB
}

func NewBookUpdateRequestRepository(db *gorm.DB) *BookUpdateRequestRepository {
	return &BookUpdateRequestRepository{db}
}

func (r *BookUpdateRequestRepository) Create(req *entities.BookUpdateRequest) error {
	return r.db.Create(req).Error
}

func (r *BookUpdateRequestRepository) FindByID(id string) (*entities.BookUpdateRequest, error) {
	var req entities.BookUpdateRequest
	err := r.db.Where("id = ?", id).First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *BookUpdateRequestRepository) FindByBookID(bookID string) ([]entities.BookUpdateRequest, error) {
	var reqs []entities.BookUpdateRequest
	err := r.db.Where("book_id = ?", bookID).Order("created_at DESC").Find(&reqs).Error
	return reqs, err
}

func (r *BookUpdateRequestRepository) FindPendingByBookID(bookID string) (*entities.BookUpdateRequest, error) {
	var req entities.BookUpdateRequest
	err := r.db.Where("book_id = ? AND status = ?", bookID, entities.BookUpdateStatusPending).First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *BookUpdateRequestRepository) FindAllPending() ([]entities.BookUpdateRequest, error) {
	var reqs []entities.BookUpdateRequest
	err := r.db.Where("status = ?", entities.BookUpdateStatusPending).Order("created_at DESC").Find(&reqs).Error
	return reqs, err
}

func (r *BookUpdateRequestRepository) Update(req *entities.BookUpdateRequest) error {
	return r.db.Save(req).Error
}

func (r *BookUpdateRequestRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&entities.BookUpdateRequest{}).Error
}
