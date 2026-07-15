package handlers

import (
	"fmt"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/cache"
	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"
)

// ClassDailyHandler handles GET /api/v1/class-daily
// Returns today's daily tasks scoped to the user's quran class items only.
type ClassDailyHandler struct {
	dailyTaskSvc    services.DailyTaskService
	dailyTaskRepo   repositories.DailyTaskRepository
	itemRepo        *repositories.ItemRepository
	juzRepo         *repositories.JuzRepository
	juzItemRepo     *repositories.JuzItemRepository
	classMemberRepo repositories.ClassMemberRepository
	classRepo       repositories.ClassRepository
	classBookRepo   repositories.ClassBookRepository
	bookRepo        repositories.BookRepository
	bookItemRepo    repositories.BookItemRepository
	cache           *cache.Cache
}

func NewClassDailyHandler(
	dailyTaskSvc services.DailyTaskService,
	dailyTaskRepo repositories.DailyTaskRepository,
	itemRepo *repositories.ItemRepository,
	juzRepo *repositories.JuzRepository,
	juzItemRepo *repositories.JuzItemRepository,
	classMemberRepo repositories.ClassMemberRepository,
	classRepo repositories.ClassRepository,
	classBookRepo repositories.ClassBookRepository,
	bookRepo repositories.BookRepository,
	bookItemRepo repositories.BookItemRepository,
	c *cache.Cache,
) *ClassDailyHandler {
	return &ClassDailyHandler{
		dailyTaskSvc:    dailyTaskSvc,
		dailyTaskRepo:   dailyTaskRepo,
		itemRepo:        itemRepo,
		juzRepo:         juzRepo,
		juzItemRepo:     juzItemRepo,
		classMemberRepo: classMemberRepo,
		classRepo:       classRepo,
		classBookRepo:   classBookRepo,
		bookRepo:        bookRepo,
		bookItemRepo:    bookItemRepo,
		cache:           c,
	}
}

// ListClassDaily godoc
// @Summary List today's class-scoped daily tasks (Quran)
// @Description Get all daily tasks for today that are scoped to the user's quran class.
// @Description Requires the user to be a member of the class.
// @Tags Daily Task
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param class_id query string true "Class UUID"
// @Param date query string false "Date override (YYYY-MM-DD, defaults to today)"
// @Param group query string false "Grouping mode: 'juz' to group by juz index"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /class-daily [get]
func (h *ClassDailyHandler) ListClassDaily(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.Error(c, fiber.StatusUnauthorized, "unauthorized", "UNAUTHORIZED", nil)
	}

	// ── 1. Parse & validate class_id ─────────────────────────────────────────
	classIDStr := c.Query("class_id", "")
	if classIDStr == "" {
		return utils.Error(c, fiber.StatusBadRequest, "class_id is required", "BAD_REQUEST", nil)
	}
	if _, err := uuid.Parse(classIDStr); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "class_id must be a valid UUID", "BAD_REQUEST", nil)
	}

	// ── 2. Verify user is a member of the class ───────────────────────────────
	isMember, err := h.classMemberRepo.IsMember(classIDStr, userID.String())
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to verify membership", "INTERNAL_ERROR", nil)
	}

	// Teacher (guru) of the class is also allowed — check ownership as fallback
	if !isMember {
		class, err := h.classRepo.FindByID(classIDStr)
		if err != nil {
			return utils.Error(c, fiber.StatusNotFound, "class not found", "NOT_FOUND", nil)
		}
		if class.GuruID != userID {
			return utils.Error(c, fiber.StatusForbidden, "you are not a member of this class", "FORBIDDEN", nil)
		}
	}

	// ── 3. Load class to verify it is a quran-type class ─────────────────────
	class, err := h.classRepo.FindByID(classIDStr)
	if err != nil {
		return utils.Error(c, fiber.StatusNotFound, "class not found", "NOT_FOUND", nil)
	}
	if class.Type != entities.ClassTypeQuran {
		return utils.Error(
			c,
			fiber.StatusBadRequest,
			"this endpoint is only for quran classes",
			"BAD_REQUEST",
			nil,
		)
	}

	// ── 4. Resolve target date ────────────────────────────────────────────────
	dateStr := c.Query("date", "")
	var now time.Time
	if dateStr != "" {
		if t, parseErr := time.ParseInLocation("2006-01-02", dateStr, config.AppLocation); parseErr == nil {
			now = t
		} else {
			now = time.Now().In(config.AppLocation)
		}
	} else {
		now = time.Now().In(config.AppLocation)
	}
	date := now.Format("2006-01-02")
	group := c.Query("group", "")

	// ── 5. Try cache ─────────────────────────────────────────────────────────
	cacheKey := fmt.Sprintf("class-daily:%s:%s:%s:%s", userID.String(), classIDStr, date, group)
	var cached []DailyTaskResponse
	if h.cache.Get(c.Context(), cacheKey, &cached) {
		return h.renderResponse(c, cached, group)
	}

	// ── 6. Collect class-scoped item IDs (via juz → juz_items) ───────────────
	juzs, err := h.juzRepo.FindByUserAndClass(userID.String(), classIDStr)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to load class juz data", "INTERNAL_ERROR", nil)
	}

	if len(juzs) == 0 {
		// User has no juz entries for this class yet → return empty result
		return utils.Success(c, fiber.StatusOK, "no class items found", []DailyTaskResponse{}, nil)
	}

	juzIDStrings := make([]string, 0, len(juzs))
	for _, j := range juzs {
		juzIDStrings = append(juzIDStrings, j.ID.String())
	}

	classItemIDStrings, err := h.juzItemRepo.FindItemIDsByJuzIDs(juzIDStrings)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to load class item IDs", "INTERNAL_ERROR", nil)
	}

	if len(classItemIDStrings) == 0 {
		return utils.Success(c, fiber.StatusOK, "no class items found", []DailyTaskResponse{}, nil)
	}

	// Build a set for O(1) membership check
	classItemSet := make(map[uuid.UUID]struct{}, len(classItemIDStrings))
	for _, idStr := range classItemIDStrings {
		if parsed, parseErr := uuid.Parse(idStr); parseErr == nil {
			classItemSet[parsed] = struct{}{}
		}
	}

	// ── 7. Load today's daily tasks then filter to class scope ────────────────
	tasks, err := h.dailyTaskRepo.ListByUserAndDate(c.Context(), userID, now)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "failed to load daily tasks", "INTERNAL_ERROR", nil)
	}

	// Auto-generate if snapshot is empty (mirrors the existing daily handler behaviour)
	if len(tasks) == 0 {
		if gen, genErr := h.dailyTaskSvc.GenerateToday(c.Context(), userID, now, 0); genErr == nil {
			tasks = gen
		}
	}

	// Keep only tasks that belong to this class AND are not yet done
	filtered := make([]entities.DailyTask, 0, len(tasks))
	for _, t := range tasks {
		if _, inClass := classItemSet[t.ItemID]; !inClass {
			continue
		}
		if t.State == "done" {
			continue
		}
		filtered = append(filtered, t)
	}

	if len(filtered) == 0 {
		return utils.Success(c, fiber.StatusOK, "no class daily tasks for today", []DailyTaskResponse{}, nil)
	}

	// ── 8. Batch-enrich with item metadata ────────────────────────────────────
	itemIDs := make([]uuid.UUID, 0, len(filtered))
	itemIDStrings := make([]string, 0, len(filtered))
	for _, t := range filtered {
		itemIDs = append(itemIDs, t.ItemID)
		itemIDStrings = append(itemIDStrings, t.ItemID.String())
	}

	itemContentMap := make(map[uuid.UUID]string)
	itemEstimateMap := make(map[uuid.UUID]int)
	itemStatusMap := make(map[uuid.UUID]string)

	if items, fetchErr := h.itemRepo.FindByIDs(itemIDs); fetchErr == nil {
		for _, item := range items {
			itemContentMap[item.ID] = item.ContentRef
			itemEstimateMap[item.ID] = item.EstimatedReviewSeconds
			itemStatusMap[item.ID] = item.Status
		}
	}

	// Juz index per item
	juzIndexMap := make(map[string]int)
	if juzResult, juzErr := h.juzItemRepo.FindJuzIndexByItemIDs(itemIDStrings); juzErr == nil {
		juzIndexMap = juzResult
	}

	// ── 9. Build response ─────────────────────────────────────────────────────
	resp := make([]DailyTaskResponse, 0, len(filtered))
	for _, t := range filtered {
		resp = append(resp, DailyTaskResponse{
			ItemID:                 t.ItemID,
			Source:                 t.Source,
			State:                  t.State,
			Status:                 itemStatusMap[t.ItemID],
			TaskDate:               t.TaskDate.Format("2006-01-02"),
			ContentRef:             itemContentMap[t.ItemID],
			JuzIndex:               juzIndexMap[t.ItemID.String()],
			EstimatedReviewSeconds: itemEstimateMap[t.ItemID],
		})
	}

	// Cache until midnight
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, config.AppLocation)
	h.cache.Set(c.Context(), cacheKey, resp, time.Until(midnight))

	return h.renderResponse(c, resp, group)
}

// renderResponse returns the response grouped by juz or as a flat list.
func (h *ClassDailyHandler) renderResponse(c *fiber.Ctx, items []DailyTaskResponse, group string) error {
	if group == "juz" {
		return utils.Success(c, fiber.StatusOK, "success", groupByJuzIndex(items), nil)
	}
	return utils.Success(c, fiber.StatusOK, "success", items, nil)
}

// groupByJuzIndex groups a slice of DailyTaskResponse by their juz_index,
// sorted ascending (juz_index 0 is placed last).
func groupByJuzIndex(items []DailyTaskResponse) []DailyTaskGroup {
	buckets := make(map[int][]DailyTaskResponse)
	for _, it := range items {
		buckets[it.JuzIndex] = append(buckets[it.JuzIndex], it)
	}

	indexes := make([]int, 0, len(buckets))
	for idx := range buckets {
		indexes = append(indexes, idx)
	}
	sort.Ints(indexes)

	var result []DailyTaskGroup
	for _, idx := range indexes {
		if idx == 0 {
			continue
		}
		result = append(result, DailyTaskGroup{JuzIndex: idx, Items: buckets[idx]})
	}
	if zero, ok := buckets[0]; ok {
		result = append(result, DailyTaskGroup{JuzIndex: 0, Items: zero})
	}
	return result
}
