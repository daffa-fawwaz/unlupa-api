package repositories

import (
	"context"

	"gorm.io/gorm"

	"hifzhun-api/pkg/entities"
)

type ItemGraduationRepository interface {
	Create(
		ctx context.Context,
		grad *entities.ItemGraduation,
	) error
}

type itemGraduationRepository struct {
	db *gorm.DB
}

func NewItemGraduationRepository(db *gorm.DB) ItemGraduationRepository {
	return &itemGraduationRepository{db: db}
}

func (r *itemGraduationRepository) Create(
	ctx context.Context,
	grad *entities.ItemGraduation,
) error {
	return r.db.WithContext(ctx).Create(grad).Error
}
