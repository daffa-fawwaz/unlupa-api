package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hifzhun-api/pkg/entities"
)

type CardRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Card, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, source string) ([]entities.Card, error)
	Create(ctx context.Context, card *entities.Card) error
	Update(ctx context.Context, card *entities.Card) error
}

type cardRepository struct {
	db *gorm.DB
}

func NewCardRepository(db *gorm.DB) CardRepository {
	return &cardRepository{db: db}
}

func (r *cardRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*entities.Card, error) {
	var card entities.Card
	if err := r.db.WithContext(ctx).
		First(&card, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &card, nil
}

func (r *cardRepository) ListByOwner(
	ctx context.Context,
	ownerID uuid.UUID,
	source string,
) ([]entities.Card, error) {
	var cards []entities.Card

	q := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID)

	if source != "" {
		q = q.Where("source = ?", source)
	}

	if err := q.Find(&cards).Error; err != nil {
		return nil, err
	}
	return cards, nil
}

func (r *cardRepository) Create(
	ctx context.Context,
	card *entities.Card,
) error {
	return r.db.WithContext(ctx).Create(card).Error
}

func (r *cardRepository) Update(
	ctx context.Context,
	card *entities.Card,
) error {
	return r.db.WithContext(ctx).Save(card).Error
}
