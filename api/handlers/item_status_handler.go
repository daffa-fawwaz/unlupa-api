package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"
)

type ItemStatusHandler struct {
	service *services.ItemStatusService
}

func NewItemStatusHandler(s *services.ItemStatusService) *ItemStatusHandler {
	return &ItemStatusHandler{service: s}
}

// StartIntervalRequest represents start interval request
type StartIntervalRequest struct {
	IntervalDays int `json:"interval_days" example:"7"`
}

// StartInterval godoc
// @Summary Start interval phase for item
// @Description Move item from menghafal to interval phase with recurring review
// @Tags Item Status
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param item_id path string true "Item ID"
// @Param request body StartIntervalRequest true "Interval days"
// @Success 200 {object} utils.SuccessResponse{data=entities.Item}
// @Failure 400 {object} utils.ErrorResponse
// @Router /items/{item_id}/start-interval [post]
func (h *ItemStatusHandler) StartInterval(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	itemID, err := uuid.Parse(c.Params("item_id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid item_id", "INVALID_PARAMETER", nil)
	}

	var req StartIntervalRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_REQUEST_BODY", nil)
	}

	item, err := h.service.StartInterval(itemID, userID, req.IntervalDays)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "START_INTERVAL_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "Item moved to interval phase with recurring review", item, nil)
}

// ReviewIntervalRequest represents interval review request
type ReviewIntervalRequest struct {
	Rating int `json:"rating" example:"3" minimum:"1" maximum:"3"` // 1=bad, 2=good, 3=perfect
}

// ReviewIntervalResponse represents interval review response
type ReviewIntervalResponse struct {
	ItemID               uuid.UUID  `json:"item_id"`
	Status               string     `json:"status"`
	Rating               int        `json:"rating"`
	RatingLabel          string     `json:"rating_label"`
	IntervalDays         int        `json:"interval_days"`
	IntervalNextReviewAt *string    `json:"interval_next_review_at"`
	ReviewCount          int        `json:"review_count"`
	ContentRef           string     `json:"content_ref"`
}

// ReviewInterval godoc
// @Summary Review an item in interval phase
// @Description Submit a review rating for an item during interval phase (1=bad, 2=good, 3=perfect)
// @Tags Item Status
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param item_id path string true "Item ID"
// @Param request body ReviewIntervalRequest true "Review rating (1=bad, 2=good, 3=perfect)"
// @Success 200 {object} utils.SuccessResponse{data=ReviewIntervalResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Router /items/{item_id}/review-interval [post]
func (h *ItemStatusHandler) ReviewInterval(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	itemID, err := uuid.Parse(c.Params("item_id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid item_id", "INVALID_PARAMETER", nil)
	}

	var req ReviewIntervalRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_REQUEST_BODY", nil)
	}

	if req.Rating < 1 || req.Rating > 3 {
		return utils.Error(c, fiber.StatusBadRequest, "Rating must be between 1 and 3 (1=bad, 2=good, 3=perfect)", "INVALID_RATING", nil)
	}

	result, err := h.service.ReviewInterval(itemID, userID, req.Rating)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "REVIEW_INTERVAL_FAILED", nil)
	}

	// Map rating to label
	ratingLabels := map[int]string{1: "bad", 2: "good", 3: "perfect"}

	var nextReviewStr *string
	if result.NextReviewAt != nil {
		s := result.NextReviewAt.Format("2006-01-02 15:04")
		nextReviewStr = &s
	}

	resp := ReviewIntervalResponse{
		ItemID:               result.Item.ID,
		Status:               result.Item.Status,
		Rating:               result.Rating,
		RatingLabel:          ratingLabels[result.Rating],
		IntervalDays:         result.Item.IntervalDays,
		IntervalNextReviewAt: nextReviewStr,
		ReviewCount:          result.Item.ReviewCount,
		ContentRef:           result.Item.ContentRef,
	}

	return utils.Success(c, fiber.StatusOK, "Interval review submitted", resp, nil)
}

// IntervalStatsResponse represents interval stats response
type IntervalStatsResponse struct {
	ItemID        uuid.UUID `json:"item_id"`
	AverageRating float64   `json:"average_rating"`
	TotalReviews  int       `json:"total_reviews"`
	Performance   string    `json:"performance"` // bad, good, perfect, no_reviews
}

// GetIntervalStats godoc
// @Summary Get interval review statistics
// @Description Get average rating and performance label for an item's interval reviews
// @Tags Item Status
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param item_id path string true "Item ID"
// @Success 200 {object} utils.SuccessResponse{data=IntervalStatsResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Router /items/{item_id}/interval-stats [get]
func (h *ItemStatusHandler) GetIntervalStats(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	itemID, err := uuid.Parse(c.Params("item_id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid item_id", "INVALID_PARAMETER", nil)
	}

	stats, err := h.service.GetIntervalReviewStats(itemID, userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "GET_STATS_FAILED", nil)
	}

	resp := IntervalStatsResponse{
		ItemID:        itemID,
		AverageRating: stats.AverageRating,
		TotalReviews:  stats.TotalReviews,
		Performance:   stats.Performance,
	}

	return utils.Success(c, fiber.StatusOK, "Interval review statistics", resp, nil)
}

// ActivateToFSRS godoc
// @Summary Activate item to FSRS phase
// @Description Move item from interval to fsrs_active phase (user decision)
// @Tags Item Status
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param item_id path string true "Item ID"
// @Success 200 {object} utils.SuccessResponse{data=entities.Item}
// @Failure 400 {object} utils.ErrorResponse
// @Router /items/{item_id}/activate-fsrs [post]
func (h *ItemStatusHandler) ActivateToFSRS(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	itemID, err := uuid.Parse(c.Params("item_id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid item_id", "INVALID_PARAMETER", nil)
	}

	item, err := h.service.ActivateToFSRS(itemID, userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "ACTIVATE_FSRS_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "Item moved to FSRS active phase", item, nil)
}

// GetByStatus godoc
// @Summary Get items by status
// @Description Get all items filtered by status
// @Tags Item Status
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string true "Status (menghafal, interval, fsrs_active, graduate)"
// @Success 200 {object} utils.SuccessResponse{data=[]entities.Item}
// @Failure 400 {object} utils.ErrorResponse
// @Router /items [get]
func (h *ItemStatusHandler) GetByStatus(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	status := c.Query("status")

	if status == "" {
		return utils.Error(c, fiber.StatusBadRequest, "status query parameter is required", "MISSING_PARAMETER", nil)
	}

	items, err := h.service.GetItemsByStatus(userID, status)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "GET_ITEMS_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "Items retrieved", items, nil)
}

// GetDeadlines godoc
// @Summary Get items with deadline reached
// @Description Get items that have reached their interval deadline
// @Tags Item Status
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=[]entities.Item}
// @Failure 500 {object} utils.ErrorResponse
// @Router /items/deadlines [get]
func (h *ItemStatusHandler) GetDeadlines(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	items, err := h.service.GetDeadlineItems(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_DEADLINES_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "Items with deadline reached", items, nil)
}

// DeactivateItem godoc
// @Summary Deactivate a book item
// @Description Move book item from fsrs_active to inactive status. Only for non-quran items.
// @Tags Item Status
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param item_id path string true "Item ID"
// @Success 200 {object} utils.SuccessResponse{data=entities.Item}
// @Failure 400 {object} utils.ErrorResponse
// @Router /items/{item_id}/deactivate [post]
func (h *ItemStatusHandler) DeactivateItem(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	itemID, err := uuid.Parse(c.Params("item_id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid item_id", "INVALID_PARAMETER", nil)
	}

	item, err := h.service.DeactivateItem(itemID, userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "DEACTIVATE_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "Item deactivated successfully", item, nil)
}

// ReactivateItem godoc
// @Summary Reactivate a book item
// @Description Move book item from inactive back to fsrs_active status. Only for non-quran items.
// @Tags Item Status
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param item_id path string true "Item ID"
// @Success 200 {object} utils.SuccessResponse{data=entities.Item}
// @Failure 400 {object} utils.ErrorResponse
// @Router /items/{item_id}/reactivate [post]
func (h *ItemStatusHandler) ReactivateItem(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	itemID, err := uuid.Parse(c.Params("item_id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid item_id", "INVALID_PARAMETER", nil)
	}

	item, err := h.service.ReactivateItem(itemID, userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "REACTIVATE_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "Item reactivated successfully", item, nil)
}
