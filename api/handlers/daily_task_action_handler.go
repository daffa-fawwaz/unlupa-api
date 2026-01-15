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
