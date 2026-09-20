package routes

import (
	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterDashboardRoutes(router fiber.Router, handler *handlers.DashboardHandler) {
	dashboard := router.Group("/dashboard", middlewares.JWTAuth())
	dashboard.Get("/stats", handler.GetStats)
}
