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

type DashboardHandler struct {
	service *services.DashboardService
	cache   *cache.Cache
}

func NewDashboardHandler(service *services.DashboardService, c *cache.Cache) *DashboardHandler {
	return &DashboardHandler{
		service: service,
		cache:   c,
	}
}

// GetStats returns aggregated dashboard metrics for the student home space
func (h *DashboardHandler) GetStats(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return fiber.ErrUnauthorized
	}

	cacheKey := fmt.Sprintf("unlupa:dashboard:stats:%s", userID.String())
	var cached services.DashboardStatsResponse
	if h.cache != nil && h.cache.Get(c.Context(), cacheKey, &cached) {
		return utils.Success(c, fiber.StatusOK, "dashboard stats fetched from cache", cached, nil)
	}

	stats, err := h.service.GetDashboardStats(c.Context(), userID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "DASHBOARD_STATS_FAILED", nil)
	}

	if h.cache != nil {
		h.cache.Set(c.Context(), cacheKey, stats, 5*time.Minute)
	}

	return utils.Success(c, fiber.StatusOK, "dashboard stats fetched successfully", stats, nil)
}
