package utils

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")

	if token == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}

	claims, err := ParseJWT(token)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}

	c.Locals("username", claims.Username)
	c.Locals("id", claims.ID)

	return c.Next()
}

func FiberContextMiddleware(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c.SetUserContext(ctx)

	return c.Next()
}
