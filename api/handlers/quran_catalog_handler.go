package handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/cache"
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"
)

type QuranCatalogHandler struct {
	service *services.QuranCatalogService
	cache   *cache.Cache
}

func NewQuranCatalogHandler(service *services.QuranCatalogService, cache *cache.Cache) *QuranCatalogHandler {
	return &QuranCatalogHandler{
		service: service,
		cache:   cache,
	}
}

func (h *QuranCatalogHandler) invalidateUserCaches(c *fiber.Ctx, userID uuid.UUID) {
	if h.cache == nil {
		return
	}
	ctx := c.Context()
	h.cache.DeleteByPattern(ctx, fmt.Sprintf("quran:juzs:%s*", userID.String()))
	h.cache.DeleteByPattern(ctx, fmt.Sprintf("quran:pages:%s*", userID.String()))
	h.cache.DeleteByPattern(ctx, fmt.Sprintf("myitems:%s:*", userID.String()))
	h.cache.Delete(ctx, fmt.Sprintf("juz:list:%s", userID.String()))
	h.cache.DeleteByPattern(ctx, fmt.Sprintf("juz:list:%s:*", userID.String()))
	h.cache.DeleteByPattern(ctx, fmt.Sprintf("daily:%s:*", userID.String()))
}

// GetJuzs godoc
// @Summary Get all 30 Juz with user progress stats
// @Description Returns static Juz 1-30 with total_pages, active_pages, mastered_pages, due_today
// @Tags Quran Catalog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=[]repositories.JuzCatalogWithUserStats}
// @Failure 401 {object} utils.ErrorResponse
// @Router /quran/juzs [get]
func (h *QuranCatalogHandler) GetJuzs(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.Error(c, fiber.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED", nil)
	}

	targetUserID := userID
	queryUserID := c.Query("user_id")
	if queryUserID != "" {
		if parsedUID, err := uuid.Parse(queryUserID); err == nil {
			targetUserID = parsedUID
		}
	}

	result, err := h.service.GetAllJuzs(c.Context(), targetUserID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_JUZS_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "Daftar Juz berhasil diambil", result, nil)
}

// GetJuzPages godoc
// @Summary Get all pages in a specific Juz with user activation/review status
// @Description Returns all Mushaf Madani pages in the Juz with activation status and review metrics
// @Tags Quran Catalog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param juzNumber path int true "Juz Number (1-30)"
// @Success 200 {object} utils.SuccessResponse{data=services.JuzPagesDetailResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /quran/juzs/{juzNumber}/pages [get]
func (h *QuranCatalogHandler) GetJuzPages(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.Error(c, fiber.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED", nil)
	}

	juzNumber, err := strconv.Atoi(c.Params("juzNumber"))
	if err != nil || juzNumber < 1 || juzNumber > 30 {
		return utils.Error(c, fiber.StatusBadRequest, "Nomor Juz tidak valid (harus 1-30)", "INVALID_PARAMETER", nil)
	}

	targetUserID := userID
	queryUserID := c.Query("user_id")
	if queryUserID != "" {
		if parsedUID, err := uuid.Parse(queryUserID); err == nil {
			targetUserID = parsedUID
		}
	}

	result, err := h.service.GetJuzPages(c.Context(), juzNumber, targetUserID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "GET_JUZ_PAGES_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "Daftar halaman Juz berhasil diambil", result, nil)
}

// ActivatePage godoc
// @Summary Activate a Quran page to user memorization collection
// @Description Atomically adds a Quran catalog page to user's items and personal Juz with initial status menghafal
// @Tags Quran Catalog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param mushafPage path int true "Mushaf Madani Page (1-604)"
// @Success 200 {object} utils.SuccessResponse{data=services.ActivatePageResult}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /quran/pages/{mushafPage}/activate [post]
func (h *QuranCatalogHandler) ActivatePage(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.Error(c, fiber.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED", nil)
	}

	mushafPage, err := strconv.Atoi(c.Params("mushafPage"))
	if err != nil || mushafPage < 1 || mushafPage > 604 {
		return utils.Error(c, fiber.StatusBadRequest, "Nomor halaman Mushaf tidak valid (harus 1-604)", "INVALID_PARAMETER", nil)
	}

	result, err := h.service.ActivatePage(c.Context(), userID, mushafPage)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "ACTIVATE_PAGE_FAILED", nil)
	}

	h.invalidateUserCaches(c, userID)

	return utils.Success(c, fiber.StatusOK, result.Message, result, nil)
}
