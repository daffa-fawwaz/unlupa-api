package middlewares

import (
	"hifzhun-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// TeacherOnly allows only teachers and admins
func TeacherOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")

		if role == nil || (role != "teacher" && role != "admin") {
			return utils.Error(
				c,
				fiber.StatusForbidden,
				"teacher or admin access only",
				"FORBIDDEN",
				nil,
			)
		}

		return c.Next()
	}
}
