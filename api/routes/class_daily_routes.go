package routes

import (
	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

// RegisterClassDailyRoutes mounts GET /api/v1/class-daily
func RegisterClassDailyRoutes(
	router fiber.Router,
	handler *handlers.ClassDailyHandler,
) {
	router.Get(
		"/class-daily",
		middlewares.JWTAuth(),
		handler.ListClassDaily,
	)
	router.Get(
		"/class-daily-book",
		middlewares.JWTAuth(),
		handler.ListClassDailyBook,
	)
}
