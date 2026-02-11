package routes

import (
	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterMyItemRoutes(
	router fiber.Router,
	myItemHandler *handlers.MyItemHandler,
) {
	myItems := router.Group(
		"/my-items",
		middlewares.JWTAuth(),
	)

	myItems.Get("/", myItemHandler.GetMyItems)
}
