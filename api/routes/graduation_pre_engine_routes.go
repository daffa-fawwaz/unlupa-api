package routes

import (
	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterGraduationPreEngineRoutes(
	router fiber.Router,
	handler *handlers.GraduationPreEngineHandler,
) {
	grad := router.Group(
		"/graduation",
		middlewares.JWTAuth(),
	)

	grad.Post("/decide", handler.Decide)
}
