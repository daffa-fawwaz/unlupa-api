package routes

import (
	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterItemStatusRoutes(
	router fiber.Router,
	handler *handlers.ItemStatusHandler,
	reviewHandler *handlers.ItemReviewHandler,
) {
	items := router.Group("/items", middlewares.JWTAuth())

	// Get items by status
	items.Get("/", handler.GetByStatus)

	// Get items that have reached deadline
	items.Get("/deadlines", handler.GetDeadlines)

	// Start interval phase
	items.Post("/:item_id/start-interval", handler.StartInterval)

	// Review item (FSRS) - auto graduate at 30 days
	items.Post("/:item_id/review", reviewHandler.ReviewItem)
}
