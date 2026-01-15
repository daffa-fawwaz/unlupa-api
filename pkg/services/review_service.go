package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/fsrs"
	"hifzhun-api/pkg/repositories"
)

type ReviewCardResult struct {
	Card         *entities.Card
	IntervalDays int
	NextReviewAt *time.Time
}

type ReviewService interface {
	ReviewCard(
		ctx context.Context,
		cardID uuid.UUID,
		rating fsrs.Rating,
		now time.Time,
	) (*ReviewCardResult, error)
}

type reviewService struct {
	db              *gorm.DB
	cardRepo        repositories.CardRepository
	reviewLogRepo   repositories.ReviewLogRepository
	reviewStateRepo repositories.ReviewStateRepository
	fsrsWeightsRepo repositories.FSRSWeightsRepository
}

func NewReviewService(
	db *gorm.DB,
	cardRepo repositories.CardRepository,
	reviewLogRepo repositories.ReviewLogRepository,
	reviewStateRepo repositories.ReviewStateRepository,
	fsrsWeightsRepo repositories.FSRSWeightsRepository,
) ReviewService {
	return &reviewService{
		db:              db,
		cardRepo:        cardRepo,
		reviewLogRepo:   reviewLogRepo,
		reviewStateRepo: reviewStateRepo,
		fsrsWeightsRepo: fsrsWeightsRepo,
	}
}

func (s *reviewService) ReviewCard(
	ctx context.Context,
	cardID uuid.UUID,
	rating fsrs.Rating,
	now time.Time,
) (*ReviewCardResult, error) {

	if rating < fsrs.Again || rating > fsrs.Easy {
		return nil, errors.New("invalid rating")
	}

	var (
		returnCard   entities.Card
		intervalDays int
		nextReviewAt *time.Time
	)

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 1️⃣ Ambil card
		card, err := s.cardRepo.GetByID(ctx, cardID)
		if err != nil {
			return err
		}

		isFirstReview := card.LastReviewAt.IsZero()
		isSameDay := !isFirstReview && now.Sub(card.LastReviewAt) < 24*time.Hour

		// 2️⃣ Initial state (LOCKED)
		if isFirstReview {
			card.Stability = 0.4
			card.Difficulty = 5.0
		}

		// 3️⃣ Ambil weights (SAFE)
		var weights fsrs.Weights
		if isFirstReview {
			weights = fsrs.DefaultWeights()
		} else {
			w, err := s.fsrsWeightsRepo.GetLatestByOwner(ctx, card.OwnerID)
			if err != nil {
				return err
			}
			raw := fsrs.NewWeights([]float64{
				w.W0, w.W1, w.W2, w.W3, w.W4, w.W5, w.W6, w.W7, w.W8,
				w.W9, w.W10, w.W11, w.W12, w.W13, w.W14, w.W15, w.W16,
			})
			weights = sanitizeWeights(raw)
		}

		// 4️⃣ State sebelum review
		prevState := fsrs.CardState{
			Stability:  card.Stability,
			Difficulty: card.Difficulty,
			LastReview: card.LastReviewAt,
		}

		var result fsrs.ReviewResult

		switch {
		case isSameDay:
			// Same-day: tidak ubah jadwal
			result = fsrs.ReviewResult{NewState: prevState}
			intervalDays = 0

			// Ambil jadwal lama
			rs, err := s.reviewStateRepo.FindByCardID(ctx, card.ID)
			if err == nil {
				nextReviewAt = rs.NextReviewAt
			}

		default:
			// Review normal
			result = fsrs.Review(prevState, rating, now, weights)

			card.Stability = result.NewState.Stability
			card.Difficulty = result.NewState.Difficulty
			card.LastReviewAt = now

			intervalDays = int(result.Interval.Hours() / 24)

			if err := s.cardRepo.Update(ctx, card); err != nil {
				return err
			}

			nr := now.Add(result.Interval)
			nextReviewAt = &nr

			reviewState := &entities.ReviewState{
				UserID:       card.OwnerID,
				ItemID:       card.ItemID,
				CardID:       card.ID,
				State:        "active",
				Stability:    card.Stability,
				NextReviewAt: nextReviewAt,
			}

			if err := s.reviewStateRepo.Upsert(ctx, reviewState); err != nil {
				return err
			}
		}

		// 7️⃣ Review log (SELALU)
		log := &entities.ReviewLog{
			ID: uuid.New(),

			UserID:     card.OwnerID,
			ItemID:     card.ItemID,
			CardID:     card.ID,
			ReviewedAt: now,
			Rating:     int(rating),

			StabilityBefore:  prevState.Stability,
			DifficultyBefore: prevState.Difficulty,

			StabilityAfter:  result.NewState.Stability,
			DifficultyAfter: result.NewState.Difficulty,

			IntervalDays: intervalDays,
		}

		if err := s.reviewLogRepo.Create(ctx, log); err != nil {
			return err
		}

		returnCard = *card
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ReviewCardResult{
		Card:         &returnCard,
		IntervalDays: intervalDays,
		NextReviewAt: nextReviewAt,
	}, nil
}

// ===============================
// SAFE WEIGHTS SANITIZER
// ===============================
func sanitizeWeights(w fsrs.Weights) fsrs.Weights {
	if len(w.W) != 17 {
		return fsrs.DefaultWeights()
	}
	if w.W[9] <= 1 {
		return fsrs.DefaultWeights()
	}
	return w
}
