package routes

import (
	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterJuzRoutes(
	router fiber.Router,
	juzHandler *handlers.JuzHandler,
	juzItemHandler *handlers.JuzItemHandler,
) {
	juz := router.Group(
		"/juz",
		middlewares.JWTAuth(), // ✅ WAJIB
	)

	juz.Post("/:index", juzHandler.Create)
	juz.Post("/:juz_id/items", juzItemHandler.Create)
}
