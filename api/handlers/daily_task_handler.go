package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/cache"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
)

// DailyTaskResponse represents daily task response
type DailyTaskResponse struct {
	ItemID     uuid.UUID `json:"item_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Source     string    `json:"source" example:"quran"`
	State      string    `json:"state" example:"pending"`
	TaskDate   string    `json:"task_date" example:"2026-02-06"` // YYYY-MM-DD
	ContentRef string    `json:"content_ref" example:"surah:78:1-5"`
	JuzIndex   int       `json:"juz_index" example:"30"`
	EstimatedReviewSeconds int `json:"estimated_review_seconds" example:"120"`
}

type DailyTaskHandler struct {
	service     services.DailyTaskService
	itemRepo    *repositories.ItemRepository
	juzItemRepo *repositories.JuzItemRepository
	cache       *cache.Cache
}

func NewDailyTaskHandler(
	service services.DailyTaskService,
	itemRepo *repositories.ItemRepository,
	juzItemRepo *repositories.JuzItemRepository,
	c *cache.Cache,
) *DailyTaskHandler {
	return &DailyTaskHandler{
		service:     service,
		itemRepo:    itemRepo,
		juzItemRepo: juzItemRepo,
		cache:       c,
	}
}

// GenerateToday godoc
// @Summary Generate today's tasks
// @Description Generate daily tasks for today
// @Tags Daily Task
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit number of tasks" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} fiber.Error
// @Failure 500 {object} fiber.Error
// @Router /daily-tasks/generate [post]
func (h *DailyTaskHandler) GenerateToday(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.ErrUnauthorized
	}

	limit, _ := strconv.Atoi(c.Query("limit", "0"))

	tasks, err := h.service.GenerateToday(
		c.Context(),
		userID,
		time.Now(),
		limit,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"task_date": time.Now().Format("2006-01-02"),
		"count":     len(tasks),
	})
}

// ListToday godoc
// @Summary List today's tasks
// @Description Get all daily tasks for today
// @Tags Daily Task
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} DailyTaskResponse
// @Failure 401 {object} fiber.Error
// @Failure 500 {object} fiber.Error
// @Router /daily-tasks/today [get]
func (h *DailyTaskHandler) ListToday(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.ErrUnauthorized
	}

	now := time.Now()
	date := now.Format("2006-01-02")

	// Try cache first
	cacheKey := fmt.Sprintf("daily:%s:%s", userID.String(), date)
	var cached []DailyTaskResponse
	if h.cache.Get(c.Context(), cacheKey, &cached) {
		return c.JSON(cached)
	}

	tasks, err := h.service.ListToday(
		c.Context(),
		userID,
		now,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Collect item IDs for batch lookup
	itemIDs := make([]uuid.UUID, 0, len(tasks))
	itemIDStrings := make([]string, 0, len(tasks))
	for _, t := range tasks {
		itemIDs = append(itemIDs, t.ItemID)
		itemIDStrings = append(itemIDStrings, t.ItemID.String())
	}

	itemMap := make(map[uuid.UUID]string)
	itemEstimateMap := make(map[uuid.UUID]int)
	if len(itemIDs) > 0 {
		items, err := h.itemRepo.FindByIDs(itemIDs)
		if err == nil {
			for _, item := range items {
				itemMap[item.ID] = item.ContentRef
				itemEstimateMap[item.ID] = item.EstimatedReviewSeconds
			}
		}
	}

	// Batch fetch juz indexes
	juzMap := make(map[string]int) // item_id string -> juz_index
	if len(itemIDStrings) > 0 {
		juzResult, err := h.juzItemRepo.FindJuzIndexByItemIDs(itemIDStrings)
		if err == nil {
			juzMap = juzResult
		}
	}

	resp := make([]DailyTaskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, DailyTaskResponse{
			ItemID:     t.ItemID,
			Source:     t.Source,
			State:      t.State,
			TaskDate:   t.TaskDate.Format("2006-01-02"),
			ContentRef: itemMap[t.ItemID],
			JuzIndex:   juzMap[t.ItemID.String()],
			EstimatedReviewSeconds: itemEstimateMap[t.ItemID],
		})
	}

	// Cache until midnight
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	ttl := time.Until(midnight)
	h.cache.Set(c.Context(), cacheKey, resp, ttl)

	return c.JSON(resp)
}
