package api

import (
	"github.com/gofiber/fiber/v2"
	_ "github.com/hm-edu/portal-common/models"
)

// whoamiHandler godoc
// @Summary whoami Endpoint
// @Description Returns the username of the logged in user
// @Tags User
// @Accept json
// @Produce json
// @Router /whoami [get]
// @Security API
// @Success 200 {string} string "Username"
// @Failure 401 {object} models.Error "Forbidden"
// @Failure 403 {object} models.Error "Unauthorized"
func (s *APIServer) whoamiHandler(c *fiber.Ctx) (err error) {
	sub := c.Locals("subject")

	return c.JSON(sub)
}
