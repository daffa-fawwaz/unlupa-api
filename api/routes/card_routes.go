package routes

import (
	"github.com/gofiber/fiber/v2"

	"hifzhun-api/api/handlers"
)

func RegisterCardRoutes(
	router fiber.Router,
	cardHandler *handlers.CardHandler,
) {
	cards := router.Group("/cards")

	// Review card (FSRS)
	cards.Post("/:id/review", cardHandler.ReviewCard)
}
