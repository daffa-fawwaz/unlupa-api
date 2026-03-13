package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/cache"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"
)

type JuzItemHandler struct {
	service     *services.HafalanService
	cache       *cache.Cache
	itemRepo    *repositories.ItemRepository
	juzItemRepo *repositories.JuzItemRepository
}

func NewJuzItemHandler(s *services.HafalanService, c *cache.Cache, itemRepo *repositories.ItemRepository, juzItemRepo *repositories.JuzItemRepository) *JuzItemHandler {
	return &JuzItemHandler{service: s, cache: c, itemRepo: itemRepo, juzItemRepo: juzItemRepo}
}

// CreateHafalanRequest represents hafalan request body
type CreateHafalanRequest struct {
	Mode         string `json:"mode" example:"surah"`                      // surah | page
	ContentRef   string `json:"content_ref" example:"surah:78:1-5"`        // surah:78:1-5 | page:582 or page:585-589
	EstimateVal  int    `json:"estimate_value,omitempty" example:"45"`     // nilai estimasi
	EstimateUnit string `json:"estimate_unit,omitempty" example:"seconds"` // seconds | minutes
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
		req.EstimateVal,
		req.EstimateUnit,
	)

	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "ADD_ITEM_FAILED", nil)
	}

	// Invalidate caches
	ctx := c.Context()
	h.cache.Delete(ctx, fmt.Sprintf("juz:list:%s", userID.String()))
	h.cache.DeleteByPattern(ctx, fmt.Sprintf("myitems:%s:*", userID.String()))

	return utils.Success(c, fiber.StatusCreated, "Hafalan added successfully", result, nil)
}

type UpdateHafalanRequest struct {
	ContentRef    string `json:"content_ref" example:"surah:78:1-5"`
	EstimateValue int    `json:"estimate_value" example:"60"`
	EstimateUnit  string `json:"estimate_unit" example:"seconds"`
}

func (h *JuzItemHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	itemID, err := uuid.Parse(c.Params("item_id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid item_id", "INVALID_PARAMETER", nil)
	}

	item, err := h.itemRepo.GetByID(itemID)
	if err != nil || item.OwnerID != userID || item.SourceType != "quran" {
		return utils.Error(c, fiber.StatusBadRequest, "Item not found or not quran", "ITEM_NOT_FOUND", nil)
	}

	var req UpdateHafalanRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_REQUEST_BODY", nil)
	}

	if req.ContentRef != "" {
		item.ContentRef = req.ContentRef
	}
	if req.EstimateValue > 0 {
		est := req.EstimateValue
		switch strings.ToLower(req.EstimateUnit) {
		case "minutes", "minute", "min", "m":
			est = est * 60
		}
		if est < 0 {
			est = 0
		}
		item.EstimatedReviewSeconds = est
	}

	if err := h.itemRepo.Update(item); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "UPDATE_FAILED", nil)
	}

	ctx := c.Context()
	h.cache.DeleteByPattern(ctx, fmt.Sprintf("myitems:%s:*", userID.String()))
	return utils.Success(c, fiber.StatusOK, "Hafalan updated successfully", item, nil)
}

func (h *JuzItemHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	itemID, err := uuid.Parse(c.Params("item_id"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid item_id", "INVALID_PARAMETER", nil)
	}
	item, err := h.itemRepo.GetByID(itemID)
	if err != nil || item.OwnerID != userID || item.SourceType != "quran" {
		return utils.Error(c, fiber.StatusBadRequest, "Item not found or not quran", "ITEM_NOT_FOUND", nil)
	}

	_ = h.juzItemRepo.DeleteByItemID(itemID.String())
	if err := h.itemRepo.DeleteByID(itemID); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "DELETE_FAILED", nil)
	}

	ctx := c.Context()
	h.cache.DeleteByPattern(ctx, fmt.Sprintf("myitems:%s:*", userID.String()))
	h.cache.DeleteByPattern(ctx, fmt.Sprintf("juz:list:%s", userID.String()))
	return utils.Success(c, fiber.StatusOK, "Hafalan deleted successfully", nil, nil)
}
