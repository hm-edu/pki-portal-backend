package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func UserFromRequest(c *fiber.Ctx) string {
	token := c.Locals("user").(*jwt.Token)
	user := token.Claims.(jwt.MapClaims)["email"].(string)
	return user
}
