package repositories

import (
	"hifzhun-api/pkg/entities"

	"gorm.io/gorm"
)

type ItemQuranRepository struct {
	db *gorm.DB
}

func NewItemQuranRepository(db *gorm.DB) *ItemQuranRepository {
	return &ItemQuranRepository{db}
}

func (r *ItemQuranRepository) Create(item *entities.Item) error {
	return r.db.Create(item).Error
}
