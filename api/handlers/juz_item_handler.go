package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"
)

type JuzItemHandler struct {
	service *services.HafalanService
}

func NewJuzItemHandler(s *services.HafalanService) *JuzItemHandler {
	return &JuzItemHandler{s}
}

// CreateHafalanRequest represents hafalan request body
type CreateHafalanRequest struct {
	Mode       string `json:"mode" example:"surah"`             // surah | page
	ContentRef string `json:"content_ref" example:"surah:78:1-5"` // surah:78:1-5 | page:582
}

// Create godoc
// @Summary Add hafalan item to juz
// @Description Add a new hafalan item (surah verses or page) to a juz
// @Tags Juz Item
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param juz_id path string true "Juz ID"
// @Param request body CreateHafalanRequest true "Hafalan request"
// @Success 201 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /juz/{juz_id}/items [post]
func (h *JuzItemHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	juzID, err := uuid.Parse(c.Params("juz_id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid juz_id parameter", "INVALID_PARAMETER", nil)
	}

	var req CreateHafalanRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_REQUEST_BODY", nil)
	}

	result, err := h.service.AddItemToJuz(
		userID,
		juzID,
		req.Mode,
		req.ContentRef,
	)

	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "ADD_ITEM_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusCreated, "Hafalan added successfully", result, nil)
}
