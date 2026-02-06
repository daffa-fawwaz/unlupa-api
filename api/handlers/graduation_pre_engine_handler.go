package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/services"
)

// GraduationDecisionRequest represents graduation decision request
type GraduationDecisionRequest struct {
	ItemID uuid.UUID `json:"item_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Action string    `json:"action" example:"graduate" enums:"graduate,freeze,reactivate"` // graduate | freeze | reactivate
	Reason string    `json:"reason" example:"Student has mastered this item"`
}

type GraduationPreEngineHandler struct {
	service services.GraduationPreEngine
}

func NewGraduationPreEngineHandler(
	service services.GraduationPreEngine,
) *GraduationPreEngineHandler {
	return &GraduationPreEngineHandler{service: service}
}

// Decide godoc
// @Summary Make graduation decision
// @Description Decide to graduate, freeze, or reactivate an item
// @Tags Graduation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body GraduationDecisionRequest true "Graduation decision"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} fiber.Error
// @Router /graduation/decide [post]
func (h *GraduationPreEngineHandler) Decide(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req GraduationDecisionRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.ErrBadRequest
	}

	if err := h.service.Decide(
		c.Context(),
		userID,
		req.ItemID,
		req.Action,
		req.Reason,
		time.Now(),
	); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"action": req.Action,
	})
}
