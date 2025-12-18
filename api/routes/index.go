package routes

import (
	"hifzhun-api/api/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	teacherReqHandler *handlers.TeacherRequestHandler,
) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	AuthRoutes(v1, authHandler)
	UserRoutes(v1, teacherReqHandler)
	AdminRoutes(v1, userHandler, teacherReqHandler)
}
