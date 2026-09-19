package routes

import (
	"hifzhun-api/api/handlers"
	"hifzhun-api/pkg/middlewares"

	"github.com/gofiber/fiber/v2"
)

func RegisterQuranPageRoutes(router fiber.Router, handler *handlers.QuranPageHandler) {
	quran := router.Group("/quran", middlewares.JWTAuth())

	quran.Get("/pages", handler.GetPages)
	quran.Post("/pages/review", handler.ReviewPage)
	quran.Get("/juz30", handler.GetJuz30)
}
