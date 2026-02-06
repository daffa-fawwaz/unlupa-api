package repositories

import (
	"hifzhun-api/pkg/entities"

	"gorm.io/gorm"
)

type JuzItemRepository struct {
	db *gorm.DB
}

func NewJuzItemRepository(db *gorm.DB) *JuzItemRepository {
	return &JuzItemRepository{db}
}

func (r *JuzItemRepository) Create(rel *entities.JuzItem) error {
	return r.db.Create(rel).Error
}
