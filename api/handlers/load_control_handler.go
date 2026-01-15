package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/services"
)

type LoadControlResponse struct {
	ItemID       uuid.UUID `json:"item_id"`
	CardID       uuid.UUID `json:"card_id"`
	State        string    `json:"state"`
	Stability    float64   `json:"stability"`
	NextReviewAt *time.Time `json:"next_review_at"`
}

type LoadControlHandler struct {
	service services.LoadControlService
}

func NewLoadControlHandler(
	service services.LoadControlService,
) *LoadControlHandler {
	return &LoadControlHandler{service: service}
}

func (h *LoadControlHandler) Today(c *fiber.Ctx) error {

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.ErrUnauthorized
	}

	limitStr := c.Query("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	items, err := h.service.SelectForToday(
		c.Context(),
		userID,
		time.Now(),
		limit,
	)
	if err != nil {
		return fiber.NewError(
			fiber.StatusInternalServerError,
			err.Error(),
		)
	}

	resp := make([]LoadControlResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, LoadControlResponse{
			ItemID:       it.ItemID,
			CardID:       it.CardID,
			State:        it.State,
			Stability:    it.Stability,
			NextReviewAt: it.NextReviewAt,
		})
	}

	return c.JSON(resp)
}
