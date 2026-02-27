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

	// Book CRUD - specific paths first
	books.Post("/", bookHandler.CreateBook)
	books.Get("/", bookHandler.GetMyBooks)
	books.Get("/published", bookHandler.GetPublishedBooks)

	// Module static paths (before dynamic /:id)
	books.Put("/modules/:id", bookHandler.UpdateModule)
	books.Delete("/modules/:id", bookHandler.DeleteModule)
	books.Post("/modules/:module_id/items", bookHandler.AddItemToModule)

	// Item static paths (before dynamic /:id)
	books.Put("/items/:id", bookHandler.UpdateItem)
	books.Delete("/items/:id", bookHandler.DeleteItem)

	// Dynamic book routes
	books.Get("/:id", bookHandler.GetBookDetail)
	books.Put("/:id", bookHandler.UpdateBook)
	books.Delete("/:id", bookHandler.DeleteBook)
	books.Post("/:id/request-publish", bookHandler.RequestPublish)
	books.Post("/:id/modules", bookHandler.AddModule)
	books.Post("/:id/items", bookHandler.AddItemToBook)

	// Memorization - start memorizing a specific book item
	books.Post("/:id/items/:item_id/start", bookHandler.StartMemorization)
}

