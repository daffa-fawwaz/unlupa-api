package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
)

type GraduationPreEngine interface {
	Decide(
		ctx context.Context,
		userID uuid.UUID,
		itemID uuid.UUID,
		action string,
		reason string,
		now time.Time,
	) error
}

type graduationPreEngine struct {
	repo repositories.ItemGraduationRepository
}

func NewGraduationPreEngine(
	repo repositories.ItemGraduationRepository,
) GraduationPreEngine {
	return &graduationPreEngine{repo: repo}
}

func (s *graduationPreEngine) Decide(
	ctx context.Context,
	userID uuid.UUID,
	itemID uuid.UUID,
	action string,
	reason string,
	now time.Time,
) error {

	switch action {
	case "graduate", "freeze", "reactivate":
		// valid
	default:
		return errors.New("invalid graduation action")
	}

	record := &entities.ItemGraduation{
		UserID:    userID,
		ItemID:    itemID,
		Action:    action,
		Reason:    reason,
		CreatedAt: now,
	}

	return s.repo.Create(ctx, record)
}
