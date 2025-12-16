package routes

import (
	"hifzhun-api/api/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authHandler *handlers.AuthHandler) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	AuthRoutes(v1, authHandler)
}
