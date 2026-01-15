package routes

import (
	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterDailyTaskRoutes(
	router fiber.Router,
	handler *handlers.DailyTaskHandler,
) {
	daily := router.Group(
		"/daily",
		middlewares.JWTAuth(), // ✅ WAJIB
	)

	daily.Post("/generate", handler.GenerateToday)
	daily.Get("", handler.ListToday)
}
