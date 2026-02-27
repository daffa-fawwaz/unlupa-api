package routes

import (
	"github.com/gofiber/fiber/v2"

	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"
)

func RegisterLoadControlRoutes(
	router fiber.Router,
	handler *handlers.LoadControlHandler,
) {
	load := router.Group(
		"/load-control",
		middlewares.JWTAuth(), // langsung di sini
	)

	load.Get("/today", handler.Today)
}
