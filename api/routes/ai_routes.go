package routes

import (
	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterAIRoutes(router fiber.Router, handler *handlers.AIHandler) {
	ai := router.Group("/ai", middlewares.JWTAuth())
	ai.Post("/generate-book", handler.GenerateBook)
	ai.Post("/generate-cards", handler.GenerateCards)
}
