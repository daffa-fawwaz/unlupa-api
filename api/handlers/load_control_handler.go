package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/services"
)

// LoadControlResponse represents load control response
type LoadControlResponse struct {
	ItemID       uuid.UUID  `json:"item_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CardID       uuid.UUID  `json:"card_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	State        string     `json:"state" example:"learning"`
	Stability    float64    `json:"stability" example:"2.5"`
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

// Today godoc
// @Summary Get today's load
// @Description Get items selected for today's review session
// @Tags Load Control
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit number of items" default(10)
// @Success 200 {array} LoadControlResponse
// @Failure 401 {object} fiber.Error
// @Failure 500 {object} fiber.Error
// @Router /load/today [get]
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
