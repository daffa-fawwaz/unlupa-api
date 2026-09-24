package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/cache"
	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/fsrs"
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"
)

type QuranPageHandler struct {
	service *services.QuranPageService
	cache   *cache.Cache
}

func NewQuranPageHandler(service *services.QuranPageService, c *cache.Cache) *QuranPageHandler {
	return &QuranPageHandler{service: service, cache: c}
}

type ReviewPageRequest struct {
	PageNumber int `json:"page_number" example:"582" minimum:"1" maximum:"604"`
	Rating     int `json:"rating" example:"3" minimum:"1" maximum:"4"` // 1=Again, 2=Hard, 3=Good, 4=Easy
}

// GetPages godoc
// @Summary Get Mushaf Madani 604 pages progress
// @Description Returns memorization and review status for all 604 pages of Mushaf Madani
// @Tags Quran Pages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /quran/pages [get]
func (h *QuranPageHandler) GetPages(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.ErrUnauthorized
	}

	targetUserID := userID
	queryUserID := c.Query("user_id")
	if queryUserID != "" {
		if parsedUID, err := uuid.Parse(queryUserID); err == nil {
			targetUserID = parsedUID
		}
	}

	pages, stats, err := h.service.GetAllPagesProgress(c.Context(), targetUserID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_PAGES_FAILED", nil)
	}

	type ResponseData struct {
		Stats *services.QuranPagesStats   `json:"stats"`
		Pages []services.QuranPageSummary `json:"pages"`
	}

	return utils.Success(c, fiber.StatusOK, "quran pages progress fetched successfully", ResponseData{
		Stats: stats,
		Pages: pages,
	}, nil)
}

// ReviewPage godoc
// @Summary Review a specific Quran page
// @Description Submit an FSRS review rating for a specific Mushaf page
// @Tags Quran Pages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ReviewPageRequest true "Page and rating"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /quran/pages/review [post]
func (h *QuranPageHandler) ReviewPage(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var req ReviewPageRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_REQUEST_BODY", nil)
	}

	if req.PageNumber < 1 || req.PageNumber > 604 {
		return utils.Error(c, fiber.StatusBadRequest, "Page number must be between 1 and 604", "INVALID_PAGE_NUMBER", nil)
	}
	if req.Rating < 1 || req.Rating > 4 {
		return utils.Error(c, fiber.StatusBadRequest, "Rating must be between 1 and 4", "INVALID_RATING", nil)
	}

	now := time.Now().In(config.AppLocation)
	progress, err := h.service.ReviewPage(c.Context(), userID, req.PageNumber, fsrs.Rating(req.Rating), now)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "REVIEW_PAGE_FAILED", nil)
	}

	// Invalidate relevant cache keys
	if h.cache != nil {
		ctx := c.Context()
		h.cache.Delete(ctx, fmt.Sprintf("unlupa:dashboard:stats:%s", userID.String()))
		h.cache.DeleteByPattern(ctx, fmt.Sprintf("quran:pages:%s*", userID.String()))
		h.cache.DeleteByPattern(ctx, fmt.Sprintf("quran:juzs:%s*", userID.String()))
		h.cache.DeleteByPattern(ctx, fmt.Sprintf("juz:list:%s*", userID.String()))
		h.cache.DeleteByPattern(ctx, fmt.Sprintf("daily:%s*", userID.String()))
		h.cache.DeleteByPattern(ctx, fmt.Sprintf("myitems:%s:*", userID.String()))
	}

	msg := fmt.Sprintf("Halaman %d berhasil dimurajaah", req.PageNumber)
	if progress.Status == "mapan" {
		msg = fmt.Sprintf("Halaman %d kini berstatus MAPAN! 🎉", req.PageNumber)
	}

	return utils.Success(c, fiber.StatusOK, msg, progress, nil)
}

// GetJuz30 godoc
// @Summary Get Juz 30 detailed progress
// @Description Returns progress and Surah checklist for Juz 30 (Surah 78-114)
// @Tags Quran Pages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /quran/juz30 [get]
func (h *QuranPageHandler) GetJuz30(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.ErrUnauthorized
	}

	result, err := h.service.GetJuz30Progress(c.Context(), userID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_JUZ30_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "juz 30 progress fetched successfully", result, nil)
}
