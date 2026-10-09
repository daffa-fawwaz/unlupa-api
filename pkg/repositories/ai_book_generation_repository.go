package repositories

import (
	"time"

	"hifzhun-api/pkg/entities"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AIBookGenerationRepository interface {
	CountWeeklyBookGenerations(userID uuid.UUID, since time.Time) (int64, error)
	LogBookGeneration(log *entities.AIBookGenerationLog) error
}

type aiBookGenerationRepository struct {
	db *gorm.DB
}

func NewAIBookGenerationRepository(db *gorm.DB) AIBookGenerationRepository {
	return &aiBookGenerationRepository{db: db}
}

func (r *aiBookGenerationRepository) CountWeeklyBookGenerations(userID uuid.UUID, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&entities.AIBookGenerationLog{}).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Count(&count).Error
	return count, err
}

func (r *aiBookGenerationRepository) LogBookGeneration(log *entities.AIBookGenerationLog) error {
	return r.db.Create(log).Error
}
