package handlers

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/cache"
	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
)

// DailyTaskResponse represents daily task response
type DailyTaskResponse struct {
	ItemID                 uuid.UUID `json:"item_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Source                 string    `json:"source" example:"quran"`
	State                  string    `json:"state" example:"pending"`
	TaskDate               string    `json:"task_date" example:"2026-02-06"` // YYYY-MM-DD
	ContentRef             string    `json:"content_ref" example:"surah:78:1-5"`
	JuzIndex               int       `json:"juz_index" example:"30"`
	EstimatedReviewSeconds int       `json:"estimated_review_seconds" example:"120"`
	BookTitle              string    `json:"book_title,omitempty" example:"Belajar Tajwid"`
}

type DailyTaskGroup struct {
	JuzIndex int                 `json:"juz_index" example:"30"`
	Items    []DailyTaskResponse `json:"items"`
}

type DailyTaskHandler struct {
	service     services.DailyTaskService
	itemRepo    *repositories.ItemRepository
	juzItemRepo *repositories.JuzItemRepository
	bookRepo    repositories.BookRepository
	cache       *cache.Cache
}

func NewDailyTaskHandler(
	service services.DailyTaskService,
	itemRepo *repositories.ItemRepository,
	juzItemRepo *repositories.JuzItemRepository,
	bookRepo repositories.BookRepository,
	c *cache.Cache,
) *DailyTaskHandler {
	return &DailyTaskHandler{
		service:     service,
		itemRepo:    itemRepo,
		juzItemRepo: juzItemRepo,
		bookRepo:    bookRepo,
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
	// Optional client date (YYYY-MM-DD) to align with device date
	dateStr := c.Query("date", "")
	var now time.Time
	if dateStr != "" {
		if t, err := time.ParseInLocation("2006-01-02", dateStr, config.AppLocation); err == nil {
			now = t
		} else {
			now = time.Now().In(config.AppLocation)
		}
	} else {
		now = time.Now().In(config.AppLocation)
	}

	tasks, err := h.service.GenerateToday(
		c.Context(),
		userID,
		now,
		limit,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Invalidate today's list cache variants (group="" and group="juz", etc).
	date := now.Format("2006-01-02")
	cachePrefix := fmt.Sprintf("daily:%s:%s:", userID.String(), date)
	h.cache.DeleteByPattern(c.Context(), cachePrefix+"*")

	return c.JSON(fiber.Map{
		"task_date": date,
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

	// Optional client date (YYYY-MM-DD)
	dateStr := c.Query("date", "")
	var now time.Time
	if dateStr != "" {
		if t, err := time.ParseInLocation("2006-01-02", dateStr, config.AppLocation); err == nil {
			now = t
		} else {
			now = time.Now().In(config.AppLocation)
		}
	} else {
		now = time.Now().In(config.AppLocation)
	}
	date := now.Format("2006-01-02")

	group := c.Query("group", "")
	// Try cache first (include group key)
	cacheKey := fmt.Sprintf("daily:%s:%s:%s", userID.String(), date, group)
	var cached []DailyTaskResponse
	if h.cache.Get(c.Context(), cacheKey, &cached) {
		if group == "juz" {
			grouped := groupByJuz(cached)
			return c.JSON(grouped)
		}
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

	// Auto-fallback: if no snapshot exists, generate now and return
	if len(tasks) == 0 {
		gen, genErr := h.service.GenerateToday(
			c.Context(),
			userID,
			now,
			0,
		)
		if genErr == nil {
			tasks = gen
			// Drop any stale cache for this date
			h.cache.Delete(c.Context(), cacheKey)
		}
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
	bookIDs := make(map[string]struct{})
	if len(itemIDs) > 0 {
		items, err := h.itemRepo.FindByIDs(itemIDs)
		if err == nil {
			for _, item := range items {
				itemMap[item.ID] = item.ContentRef
				itemEstimateMap[item.ID] = item.EstimatedReviewSeconds
				// Collect book IDs from content_ref "book:{book_id}:item:{book_item_id}"
				if len(item.ContentRef) > 5 && item.ContentRef[:5] == "book:" {
					parts := make([]string, 0, 4)
					start := 0
					for i := 0; i < len(item.ContentRef); i++ {
						if item.ContentRef[i] == ':' {
							parts = append(parts, item.ContentRef[start:i])
							start = i + 1
						}
					}
					parts = append(parts, item.ContentRef[start:])
					if len(parts) >= 2 && parts[0] == "book" {
						bid := parts[1]
						if bid != "" {
							bookIDs[bid] = struct{}{}
						}
					}
				}
			}
		}
	}

	// Fetch book titles
	bookTitleByID := make(map[string]string)
	for bid := range bookIDs {
		if h.bookRepo != nil {
			if book, err := h.bookRepo.FindByID(bid); err == nil {
				bookTitleByID[bid] = book.Title
			}
		}
	}

	// Map item -> book title
	bookTitleByItem := make(map[uuid.UUID]string)
	for id, ref := range itemMap {
		if len(ref) > 5 && ref[:5] == "book:" {
			parts := make([]string, 0, 4)
			start := 0
			for i := 0; i < len(ref); i++ {
				if ref[i] == ':' {
					parts = append(parts, ref[start:i])
					start = i + 1
				}
			}
			parts = append(parts, ref[start:])
			if len(parts) >= 2 && parts[0] == "book" {
				bid := parts[1]
				if title, ok := bookTitleByID[bid]; ok {
					bookTitleByItem[id] = title
				}
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
			ItemID:                 t.ItemID,
			Source:                 t.Source,
			State:                  t.State,
			TaskDate:               t.TaskDate.Format("2006-01-02"),
			ContentRef:             itemMap[t.ItemID],
			JuzIndex:               juzMap[t.ItemID.String()],
			EstimatedReviewSeconds: itemEstimateMap[t.ItemID],
			BookTitle:              bookTitleByItem[t.ItemID],
		})
	}

	// Cache until midnight of the provided date (app timezone)
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, config.AppLocation)
	ttl := time.Until(midnight)
	h.cache.Set(c.Context(), cacheKey, resp, ttl)

	if group == "juz" {
		grouped := groupByJuz(resp)
		return c.JSON(grouped)
	}

	return c.JSON(resp)
}

func groupByJuz(items []DailyTaskResponse) []DailyTaskGroup {
	buckets := make(map[int][]DailyTaskResponse)
	for _, it := range items {
		buckets[it.JuzIndex] = append(buckets[it.JuzIndex], it)
	}
	// order by juz_index ascending; push 0 to the end
	indexes := make([]int, 0, len(buckets))
	for idx := range buckets {
		indexes = append(indexes, idx)
	}
	sort.Ints(indexes)
	var result []DailyTaskGroup
	// collect non-zero
	for _, idx := range indexes {
		if idx == 0 {
			continue
		}
		result = append(result, DailyTaskGroup{JuzIndex: idx, Items: buckets[idx]})
	}
	// append zero group last if exists
	if items0, ok := buckets[0]; ok {
		result = append(result, DailyTaskGroup{JuzIndex: 0, Items: items0})
	}
	return result
}
