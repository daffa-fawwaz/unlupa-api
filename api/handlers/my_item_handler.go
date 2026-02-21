package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/cache"
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"
)

type MyItemHandler struct {
	service *services.MyItemService
	cache   *cache.Cache
}

func NewMyItemHandler(service *services.MyItemService, c *cache.Cache) *MyItemHandler {
	return &MyItemHandler{service: service, cache: c}
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
	if filterType == "" {
		filterType = "all"
	}

	// Try cache first
	cacheKey := fmt.Sprintf("myitems:%s:%s", userID.String(), filterType)
	switch filterType {
	case "quran":
		var cached *services.MyItemsQuranResponse
		if h.cache.Get(c.Context(), cacheKey, &cached) {
			return utils.Success(c, fiber.StatusOK, "quran items fetched successfully", cached, nil)
		}
		result, err := h.service.GetMyQuranItems(userID)
		if err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_ITEMS_FAILED", nil)
		}
		h.cache.Set(c.Context(), cacheKey, result, 10*time.Minute)
		return utils.Success(c, fiber.StatusOK, "quran items fetched successfully", result, nil)

	case "book":
		var cached *services.MyItemsBookResponse
		if h.cache.Get(c.Context(), cacheKey, &cached) {
			return utils.Success(c, fiber.StatusOK, "book items fetched successfully", cached, nil)
		}
		result, err := h.service.GetMyBookItems(userID)
		if err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_ITEMS_FAILED", nil)
		}
		h.cache.Set(c.Context(), cacheKey, result, 10*time.Minute)
		return utils.Success(c, fiber.StatusOK, "book items fetched successfully", result, nil)

	default:
		// Return both quran and book
		type allResult struct {
			Quran *services.MyItemsQuranResponse `json:"quran"`
			Book  *services.MyItemsBookResponse  `json:"book"`
		}
		var cached allResult
		if h.cache.Get(c.Context(), cacheKey, &cached) {
			return utils.Success(c, fiber.StatusOK, "items fetched successfully", cached, nil)
		}

		quranResult, err := h.service.GetMyQuranItems(userID)
		if err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_ITEMS_FAILED", nil)
		}

		bookResult, err := h.service.GetMyBookItems(userID)
		if err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_ITEMS_FAILED", nil)
		}

		combined := allResult{Quran: quranResult, Book: bookResult}
		h.cache.Set(c.Context(), cacheKey, combined, 10*time.Minute)

		return utils.Success(c, fiber.StatusOK, "items fetched successfully", combined, nil)
	}
}

