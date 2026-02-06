package routes

import (
	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterBookRoutes(
	router fiber.Router,
	bookHandler *handlers.BookHandler,
) {
	books := router.Group(
		"/books",
		middlewares.JWTAuth(),
	)

	// Book CRUD
	books.Post("/", bookHandler.CreateBook)
	books.Get("/", bookHandler.GetMyBooks)
	books.Get("/published", bookHandler.GetPublishedBooks)
	books.Get("/:id", bookHandler.GetBookDetail)
	books.Put("/:id", bookHandler.UpdateBook)
	books.Delete("/:id", bookHandler.DeleteBook)

	// Publish workflow
	books.Post("/:id/request-publish", bookHandler.RequestPublish)

	// Module CRUD
	books.Post("/:book_id/modules", bookHandler.AddModule)
	books.Put("/modules/:id", bookHandler.UpdateModule)
	books.Delete("/modules/:id", bookHandler.DeleteModule)

	// Item CRUD
	books.Post("/:book_id/items", bookHandler.AddItemToBook)
	books.Post("/modules/:module_id/items", bookHandler.AddItemToModule)
	books.Put("/items/:id", bookHandler.UpdateItem)
	books.Delete("/items/:id", bookHandler.DeleteItem)
}
