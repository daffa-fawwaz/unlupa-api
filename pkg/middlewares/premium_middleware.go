package middlewares

import "github.com/gofiber/fiber/v2"

func PremiumOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// nanti isi
		return c.Next()
	}
}