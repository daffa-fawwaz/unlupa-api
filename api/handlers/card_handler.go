package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/fsrs"
	"hifzhun-api/pkg/services"
)

// ReviewCardRequest represents review card request body
// @Description Review card request with rating
type ReviewCardRequest struct {
	Rating int `json:"rating" example:"3" minimum:"1" maximum:"4"` // 1=Again, 2=Hard, 3=Good, 4=Easy
}

// ReviewCardResponse represents review card response
// @Description Response after reviewing a card
type ReviewCardResponse struct {
	CardID       uuid.UUID  `json:"card_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Stability    float64    `json:"stability" example:"5.5"`
	Difficulty   float64    `json:"difficulty" example:"3.2"`
	NextInterval int        `json:"next_interval_days" example:"7"`
	NextReviewAt *time.Time `json:"next_review_at"`
	LastReviewAt time.Time  `json:"last_review_at"`
}

type CardHandler struct {
	reviewService services.ReviewService
}

func NewCardHandler(
	reviewService services.ReviewService,
) *CardHandler {
	return &CardHandler{
		reviewService: reviewService,
	}
}

// ReviewCard godoc
// @Summary Review a card
// @Description Submit a review rating for a flashcard (FSRS algorithm)
// @Tags Card
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Card ID"
// @Param request body ReviewCardRequest true "Review rating (1=Again, 2=Hard, 3=Good, 4=Easy)"
// @Success 200 {object} ReviewCardResponse
// @Failure 400 {object} fiber.Error
// @Failure 500 {object} fiber.Error
// @Router /cards/{id}/review [post]
func (h *CardHandler) ReviewCard(c *fiber.Ctx) error {
	// ===============================
	// 1. PARSE CARD ID (PATH)
	// ===============================
	cardID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"invalid card id",
		)
	}

	// ===============================
	// 2. PARSE BODY
	// ===============================
	var req ReviewCardRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"invalid request body",
		)
	}

	// ===============================
	// 3. VALIDASI RATING RINGAN
	// ===============================
	if req.Rating < 1 || req.Rating > 4 {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"rating must be between 1 and 4",
		)
	}

	// ===============================
	// 4. PANGGIL SERVICE
	// ===============================
	result, err := h.reviewService.ReviewCard(
		c.Context(),
		cardID,
		fsrs.Rating(req.Rating),
		time.Now(),
	)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			err.Error(),
		)
	}

	// ===============================
	// 5. RESPONSE
	// ===============================
	resp := ReviewCardResponse{
		CardID:       result.Card.ID,
		Stability:    result.Card.Stability,
		Difficulty:   result.Card.Difficulty,
		NextInterval: result.IntervalDays,
		NextReviewAt: result.NextReviewAt,
		LastReviewAt: result.Card.LastReviewAt,
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
