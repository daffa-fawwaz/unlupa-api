package routes

import (
	"time"

	"hifzhun-api/api/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func AuthRoutes(router fiber.Router, authHandler *handlers.AuthHandler) {
	auth := router.Group("/auth")

	// Rate limit: max 5 login attempts per minute per IP
	loginLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() + ":login"
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"status":  false,
				"message": "too many login attempts, please try again later",
				"code":    "RATE_LIMITED",
			})
		},
	})

	// Rate limit: max 3 register attempts per minute per IP
	registerLimiter := limiter.New(limiter.Config{
		Max:        3,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() + ":register"
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"status":  false,
				"message": "too many register attempts, please try again later",
				"code":    "RATE_LIMITED",
			})
		},
	})

	auth.Post("/register", registerLimiter, authHandler.Register)
	auth.Post("/login", loginLimiter, authHandler.Login)
}
