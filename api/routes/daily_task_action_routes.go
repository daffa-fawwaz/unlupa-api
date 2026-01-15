package routes

import (
	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)


func RegisterDailyTaskActionRoutes(
	router fiber.Router,
	handler *handlers.DailyTaskActionHandler,
) {
	daily := router.Group("/daily", middlewares.JWTAuth())

	daily.Post("/:card_id/done", handler.MarkDone)
	daily.Post("/:card_id/skip", handler.MarkSkipped)
}
