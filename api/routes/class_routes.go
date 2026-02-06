package routes

import (
	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterClassRoutes(
	router fiber.Router,
	classHandler *handlers.ClassHandler,
) {
	classes := router.Group(
		"/classes",
		middlewares.JWTAuth(),
	)

	// ==================== STUDENT ENDPOINTS ====================
	// These can be accessed by any authenticated user (student, teacher, admin)
	classes.Post("/join", classHandler.JoinClass)
	classes.Get("/joined", classHandler.GetMyJoinedClasses)
	classes.Delete("/:id/leave", classHandler.LeaveClass)

	// ==================== TEACHER/ADMIN ENDPOINTS ====================
	// These can only be accessed by teachers and admins
	teacher := classes.Group("", middlewares.TeacherOnly())

	// Class CRUD (Teacher only)
	teacher.Post("/", classHandler.CreateClass)
	teacher.Get("/", classHandler.GetMyClasses)
	teacher.Put("/:id", classHandler.UpdateClass)
	teacher.Delete("/:id", classHandler.DeleteClass)

	// Class books management (Teacher only)
	teacher.Post("/:id/books", classHandler.AddBookToClass)
	teacher.Delete("/:id/books/:book_id", classHandler.RemoveBookFromClass)

	// Class members & progress (Teacher only)
	teacher.Get("/:id/members", classHandler.GetClassMembers)
	teacher.Get("/:id/progress", classHandler.GetStudentProgress)

	// ==================== SHARED ENDPOINTS ====================
	// These can be accessed by members of the class (student/teacher)
	classes.Get("/:id", classHandler.GetClassDetail)
	classes.Get("/:id/books", classHandler.GetClassBooks)
}

