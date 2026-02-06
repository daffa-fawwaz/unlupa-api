package routes

import (
	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

func AdminRoutes(
	router fiber.Router,
	userHandler *handlers.UserHandler,
	teacherReqHandler *handlers.TeacherRequestHandler,
	bookHandler *handlers.BookHandler,
) {
	admin := router.Group(
		"/admin",
		middlewares.JWTAuth(),
		middlewares.AdminOnly(),
	)

	admin.Get("/teacher-requests", teacherReqHandler.GetPendingRequests)
	admin.Post("/teacher-requests/:id/approve", teacherReqHandler.ApproveRequest)
	admin.Post("/teacher-requests/:id/reject", teacherReqHandler.RejectRequest)

	admin.Get("/users", userHandler.GetAllUsers)
	admin.Post("/users/:id/activate", userHandler.ActivateUser)
	admin.Post("/users/:id/deactivate", userHandler.DeactivateUser)

	// Book approval endpoints
	admin.Get("/books/pending", bookHandler.GetPendingBooks)
	admin.Post("/books/:id/approve", bookHandler.ApproveBook)
	admin.Post("/books/:id/reject", bookHandler.RejectBook)
}

