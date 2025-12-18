package routes

import (
	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

func UserRoutes(router fiber.Router, teacherReqHandler *handlers.TeacherRequestHandler) {
	user := router.Group(
		"/user",
		middlewares.JWTAuth(),
	)

	user.Post("/teacher-request", teacherReqHandler.RequestTeacher)
	user.Get("/teacher-request", teacherReqHandler.GetMyRequest)
}
