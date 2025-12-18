package routes

import (
	"hifzhun-api/api/handlers"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(router fiber.Router, authHandler *handlers.AuthHandler) {
	auth := router.Group("/auth")

	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
}
