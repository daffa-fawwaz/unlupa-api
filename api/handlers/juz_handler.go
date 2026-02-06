package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"
)

type JuzHandler struct {
	service *services.HafalanService
}

func NewJuzHandler(s *services.HafalanService) *JuzHandler {
	return &JuzHandler{s}
}

// Create godoc
// @Summary Create juz for hafalan
// @Description Create a new juz entry for Quran memorization
// @Tags Juz
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param index path int true "Juz index (1-30)"
// @Success 201 {object} utils.SuccessResponse{data=entities.Juz}
// @Failure 400 {object} utils.ErrorResponse
// @Router /juz/{index} [post]
func (h *JuzHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	juzIndex, err := strconv.Atoi(c.Params("index"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid juz index parameter", "INVALID_PARAMETER", nil)
	}

	juz, err := h.service.CreateJuz(userID, juzIndex)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "CREATE_JUZ_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusCreated, "Juz created successfully", juz, nil)
}
