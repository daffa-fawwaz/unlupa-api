package repositories

import (
	"hifzhun-api/pkg/entities"

	"gorm.io/gorm"
)

type JuzRepository struct {
	db *gorm.DB
}

func NewJuzRepository(db *gorm.DB) *JuzRepository {
	return &JuzRepository{db}
}

func (r *JuzRepository) Create(juz *entities.Juz) error {
	return r.db.Create(juz).Error
}

func (r *JuzRepository) FindByUserAndIndex(userID string, index int) (*entities.Juz, error) {
	var juz entities.Juz
	err := r.db.
		Where("user_id = ? AND index = ?", userID, index).
		First(&juz).Error
	return &juz, err
}
