package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/fsrs"
	"hifzhun-api/pkg/services"
)

type ReviewCardRequest struct {
	Rating int `json:"rating"` // 1..4
}

type ReviewCardResponse struct {
	CardID       uuid.UUID  `json:"card_id"`
	Stability    float64    `json:"stability"`
	Difficulty   float64    `json:"difficulty"`
	NextInterval int        `json:"next_interval_days"`
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
	NextReviewAt: result.NextReviewAt, // ✅ INI
	LastReviewAt: result.Card.LastReviewAt,
}

	return c.Status(fiber.StatusOK).JSON(resp)
}


