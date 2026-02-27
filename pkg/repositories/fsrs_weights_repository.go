package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hifzhun-api/pkg/entities"
)

type FSRSWeightsRepository interface {
	GetLatestByOwner(ctx context.Context, ownerID uuid.UUID) (*entities.FSRSWeights, error)
	Create(ctx context.Context, weights *entities.FSRSWeights) error
}

type fsrsWeightsRepository struct {
	db *gorm.DB
}

func NewFSRSWeightsRepository(db *gorm.DB) FSRSWeightsRepository {
	return &fsrsWeightsRepository{db: db}
}

func (r *fsrsWeightsRepository) GetLatestByOwner(
	ctx context.Context,
	ownerID uuid.UUID,
) (*entities.FSRSWeights, error) {
	var w entities.FSRSWeights

	if err := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Order("created_at desc").
		First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *fsrsWeightsRepository) Create(
	ctx context.Context,
	weights *entities.FSRSWeights,
) error {
	return r.db.WithContext(ctx).Create(weights).Error
}
