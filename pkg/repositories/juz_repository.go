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

// FindByUser returns all juz entries for a user, ordered by index
func (r *JuzRepository) FindByUser(userID string) ([]entities.Juz, error) {
	var juzs []entities.Juz
	err := r.db.
		Where("user_id = ?", userID).
		Order("\"index\" ASC").
		Find(&juzs).Error
	return juzs, err
}

// FindActiveByUser returns active juz for a user ordered by queue order then index
func (r *JuzRepository) FindActiveByUser(userID string) ([]entities.Juz, error) {
	var juzs []entities.Juz
	err := r.db.
		Where("user_id = ? AND is_active = TRUE", userID).
		Order("\"index\" ASC").
		Find(&juzs).Error
	return juzs, err
}

func (r *JuzRepository) Update(juz *entities.Juz) error {
	return r.db.Save(juz).Error
}

// SetActiveByIndex toggles is_active for a user's juz by index
func (r *JuzRepository) SetActiveByIndex(userID string, index int, active bool) error {
	return r.db.
		Model(&entities.Juz{}).
		Where("user_id = ? AND index = ?", userID, index).
		Update("is_active", active).
		Error
}

// NOTE: explicit order management removed; rotation uses index of active juzs.
