package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/services"
)

type DailyTaskResponse struct {
	ItemID   uuid.UUID `json:"item_id"`
	CardID   uuid.UUID `json:"card_id"`
	Source   string    `json:"source"`
	State    string    `json:"state"`
	TaskDate string    `json:"task_date"` // YYYY-MM-DD
}

type DailyTaskHandler struct {
	service services.DailyTaskService
}

func NewDailyTaskHandler(service services.DailyTaskService) *DailyTaskHandler {
	return &DailyTaskHandler{service: service}
}

// POST /daily-tasks/generate?limit=10
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

// GET /daily-tasks/today
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
