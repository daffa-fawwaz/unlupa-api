package repositories

import (
	"errors"
	"hifzhun-api/pkg/entities"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeacherRequestRepository interface {
	Create(req *entities.TeacherRequest) error
	FindByID(id string) (*entities.TeacherRequest, error)
	FindByUserID(userID uuid.UUID) (*entities.TeacherRequest, error)
	GetPendingRequests() ([]entities.TeacherRequest, error)
	UpdateStatus(id string, status string) error
}

type teacherRequestRepository struct {
	db *gorm.DB
}

func NewTeacherRequestRepository(db *gorm.DB) TeacherRequestRepository {
	return &teacherRequestRepository{db}
}

func (r *teacherRequestRepository) Create(req *entities.TeacherRequest) error {
	var existing entities.TeacherRequest
	err := r.db.Where("user_id = ? AND status = ?", req.UserID, "pending").First(&existing).Error
	if err == nil {
		return errors.New("you already have a pending teacher request")
	}

	return r.db.Create(req).Error
}

func (r *teacherRequestRepository) FindByID(id string) (*entities.TeacherRequest, error) {
	var req entities.TeacherRequest
	err := r.db.Preload("User").First(&req, "id = ?", id).Error
	return &req, err
}

func (r *teacherRequestRepository) FindByUserID(userID uuid.UUID) (*entities.TeacherRequest, error) {
	var req entities.TeacherRequest
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").First(&req).Error
	return &req, err
}

func (r *teacherRequestRepository) GetPendingRequests() ([]entities.TeacherRequest, error) {
	var requests []entities.TeacherRequest
	err := r.db.
		Preload("User").
		Where("status = ?", "pending").
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

func (r *teacherRequestRepository) UpdateStatus(id string, status string) error {
	return r.db.Model(&entities.TeacherRequest{}).Where("id = ?", id).Update("status", status).Error
}
