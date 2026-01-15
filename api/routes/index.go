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
	loadControlHandler *handlers.LoadControlHandler,
	cardHandler *handlers.CardHandler,
	dailyTaskHandler *handlers.DailyTaskHandler,
	dailyTaskActionHandler *handlers.DailyTaskActionHandler,
) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	AuthRoutes(v1, authHandler)
	UserRoutes(v1, teacherReqHandler)
	AdminRoutes(v1, userHandler, teacherReqHandler)
	RegisterLoadControlRoutes(v1, loadControlHandler)
	RegisterCardRoutes(v1, cardHandler)
	RegisterDailyTaskRoutes(v1, dailyTaskHandler)
	RegisterDailyTaskActionRoutes(v1, dailyTaskActionHandler)
}
