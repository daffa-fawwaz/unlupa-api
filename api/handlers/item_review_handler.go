package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/fsrs"
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"
)

type ItemReviewHandler struct {
	service *services.ItemReviewService
}

func NewItemReviewHandler(s *services.ItemReviewService) *ItemReviewHandler {
	return &ItemReviewHandler{service: s}
}

// ReviewItemRequest represents item review request
type ReviewItemRequest struct {
	Rating int `json:"rating" example:"3" minimum:"1" maximum:"4"` // 1=Again, 2=Hard, 3=Good, 4=Easy
}

// ReviewItemResponse represents item review response
type ReviewItemResponse struct {
	ItemID       uuid.UUID  `json:"item_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status       string     `json:"status" example:"fsrs_active"`
	Stability    float64    `json:"stability" example:"5.5"`
	Difficulty   float64    `json:"difficulty" example:"3.2"`
	IntervalDays int        `json:"next_interval_days" example:"7"`
	NextReviewAt *time.Time `json:"next_review_at"`
	Graduated    bool       `json:"graduated" example:"false"`
}

// ReviewItem godoc
// @Summary Review an item
// @Description Submit a review rating for a hafalan item
// @Tags Item Review
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param item_id path string true "Item ID"
// @Param request body ReviewItemRequest true "Review rating"
// @Success 200 {object} utils.SuccessResponse{data=ReviewItemResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Router /items/{item_id}/review [post]
func (h *ItemReviewHandler) ReviewItem(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	itemID, err := uuid.Parse(c.Params("item_id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid item_id", "INVALID_PARAMETER", nil)
	}

	var req ReviewItemRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_REQUEST_BODY", nil)
	}

	if req.Rating < 1 || req.Rating > 4 {
		return utils.Error(c, fiber.StatusBadRequest, "Rating must be between 1 and 4", "INVALID_RATING", nil)
	}

	result, err := h.service.ReviewItem(
		userID,
		itemID,
		fsrs.Rating(req.Rating),
		time.Now(),
	)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "REVIEW_FAILED", nil)
	}

	message := "Item reviewed"
	if result.Graduated {
		message = "Item graduated! 🎉"
	}

	resp := ReviewItemResponse{
		ItemID:       result.Item.ID,
		Status:       result.Item.Status,
		Stability:    result.Item.Stability,
		Difficulty:   result.Item.Difficulty,
		IntervalDays: result.IntervalDays,
		NextReviewAt: result.NextReviewAt,
		Graduated:    result.Graduated,
	}

	return utils.Success(c, fiber.StatusOK, message, resp, nil)
}
