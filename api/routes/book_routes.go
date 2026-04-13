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
	books.Get("/published/:id", bookHandler.GetPublishedBookDetail)
	books.Post("/published/:id/add-to-my-books", bookHandler.AddPublishedBookToMyBook)
	books.Post("/published/:id/copy-to-draft", bookHandler.CopyPublishedBookToDraft)
	
	// My Book Collection
	books.Get("/my-collection", bookHandler.GetMyBookCollection)
	books.Delete("/my-collection/:id", bookHandler.RemoveFromMyBookCollection)
	
	books.Get("/:id/tree", bookHandler.GetBookTree)
	books.Post("/:id/request-update", bookHandler.RequestBookUpdate)
	books.Get("/:id/update-requests", bookHandler.GetBookUpdateRequests)

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
