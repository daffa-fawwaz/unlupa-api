package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/services"
)

// DailyTaskResponse represents daily task response
type DailyTaskResponse struct {
	ItemID   uuid.UUID `json:"item_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CardID   uuid.UUID `json:"card_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Source   string    `json:"source" example:"quran"`
	State    string    `json:"state" example:"pending"`
	TaskDate string    `json:"task_date" example:"2026-02-06"` // YYYY-MM-DD
}

type DailyTaskHandler struct {
	service services.DailyTaskService
}

func NewDailyTaskHandler(service services.DailyTaskService) *DailyTaskHandler {
	return &DailyTaskHandler{service: service}
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

	tasks, err := h.service.ListToday(
		c.Context(),
		userID,
		now,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	resp := make([]DailyTaskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, DailyTaskResponse{
			ItemID:   t.ItemID,
			CardID:   t.CardID,
			Source:   t.Source,
			State:    t.State,
			TaskDate: t.TaskDate.Format("2006-01-02"),
		})
	}

	return c.JSON(resp)
}
