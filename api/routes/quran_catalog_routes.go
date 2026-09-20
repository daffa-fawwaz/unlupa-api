package routes

import (
	"github.com/gofiber/fiber/v2"

	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"
)

func RegisterQuranCatalogRoutes(router fiber.Router, handler *handlers.QuranCatalogHandler) {
	quran := router.Group("/quran", middlewares.JWTAuth())

	quran.Get("/juzs", handler.GetJuzs)
	quran.Get("/juzs/:juzNumber/pages", handler.GetJuzPages)
	quran.Post("/pages/:mushafPage/activate", handler.ActivatePage)
}
