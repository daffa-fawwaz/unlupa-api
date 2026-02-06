package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/services"
)

type DailyTaskActionHandler struct {
	service services.DailyTaskActionService
}

func NewDailyTaskActionHandler(
	service services.DailyTaskActionService,
) *DailyTaskActionHandler {
	return &DailyTaskActionHandler{service: service}
}

// MarkDone godoc
// @Summary Mark task as done
// @Description Mark a daily task as completed
// @Tags Daily Task
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param card_id path string true "Card ID"
// @Success 204 "No Content"
// @Failure 400 {object} fiber.Error
// @Failure 409 {object} fiber.Error
// @Router /daily-tasks/{card_id}/done [post]
func (h *DailyTaskActionHandler) MarkDone(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	cardID, err := uuid.Parse(c.Params("card_id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	if err := h.service.MarkDone(
		c.Context(),
		userID,
		cardID,
		time.Now(),
	); err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// MarkSkipped godoc
// @Summary Mark task as skipped
// @Description Mark a daily task as skipped
// @Tags Daily Task
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param card_id path string true "Card ID"
// @Success 204 "No Content"
// @Failure 400 {object} fiber.Error
// @Failure 409 {object} fiber.Error
// @Router /daily-tasks/{card_id}/skip [post]
func (h *DailyTaskActionHandler) MarkSkipped(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	cardID, err := uuid.Parse(c.Params("card_id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	if err := h.service.MarkSkipped(
		c.Context(),
		userID,
		cardID,
		time.Now(),
	); err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}
