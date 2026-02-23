package middlewares

import (
	"os"
	"strings"

	"hifzhun-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func JWTAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.Error(
				c,
				fiber.StatusUnauthorized,
				"missing authorization header",
				"UNAUTHORIZED",
				nil,
			)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return utils.Error(
				c,
				fiber.StatusUnauthorized,
				"invalid authorization format",
				"UNAUTHORIZED",
				nil,
			)
		}

		tokenStr := parts[1]
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			panic("JWT_SECRET environment variable is not set")
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return utils.Error(
				c,
				fiber.StatusUnauthorized,
				"invalid or expired token",
				"UNAUTHORIZED",
				nil,
			)
		}

		claims := token.Claims.(jwt.MapClaims)

		// parse user_id string to uuid.UUID
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return utils.Error(
				c,
				fiber.StatusUnauthorized,
				"invalid user_id in token",
				"UNAUTHORIZED",
				nil,
			)
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return utils.Error(
				c,
				fiber.StatusUnauthorized,
				"invalid user_id format",
				"UNAUTHORIZED",
				nil,
			)
		}

		// simpan ke context
		c.Locals("user_id", userID)
		c.Locals("email", claims["email"])
		c.Locals("role", claims["role"])

		return c.Next()
	}
}
