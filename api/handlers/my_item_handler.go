package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"
)

type MyItemHandler struct {
	service *services.MyItemService
}

func NewMyItemHandler(service *services.MyItemService) *MyItemHandler {
	return &MyItemHandler{service: service}
}

// GetMyItems godoc
// @Summary Get my items
// @Description Get all user's memorization items grouped by juz (quran) or book (book)
// @Tags My Items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string false "Filter by type: quran or book" Enums(quran, book)
// @Success 200 {object} utils.SuccessResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /my-items [get]
func (h *MyItemHandler) GetMyItems(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.ErrUnauthorized
	}

	filterType := c.Query("type", "")

	switch filterType {
	case "quran":
		result, err := h.service.GetMyQuranItems(userID)
		if err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_ITEMS_FAILED", nil)
		}
		return utils.Success(c, fiber.StatusOK, "quran items fetched successfully", result, nil)

	case "book":
		result, err := h.service.GetMyBookItems(userID)
		if err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_ITEMS_FAILED", nil)
		}
		return utils.Success(c, fiber.StatusOK, "book items fetched successfully", result, nil)

	default:
		// Return both quran and book
		quranResult, err := h.service.GetMyQuranItems(userID)
		if err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_ITEMS_FAILED", nil)
		}

		bookResult, err := h.service.GetMyBookItems(userID)
		if err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_ITEMS_FAILED", nil)
		}

		return utils.Success(c, fiber.StatusOK, "items fetched successfully", fiber.Map{
			"quran": quranResult,
			"book":  bookResult,
		}, nil)
	}
}
