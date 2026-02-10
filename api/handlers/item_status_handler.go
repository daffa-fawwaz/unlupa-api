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
// @Description Move item from menghafal to interval phase
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

	return utils.Success(c, fiber.StatusOK, "Item moved to interval phase", item, nil)
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

