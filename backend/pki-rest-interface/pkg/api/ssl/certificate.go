package ssl

import (
	"net/http"

	"github.com/hm-edu/pki-rest-interface/pkg/model"
	"github.com/labstack/echo/v4"
)

// List godoc
// @Summary SSL List Endpoint
// @Tags SSL
// @Accept json
// @Produce json
// @Router /ssl/ [get]
// @Security API
// @Success 200 {object} model.SSL "Certificates"
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) List(c echo.Context) error {
	return c.JSON(http.StatusOK, model.SSL{})
}

// Revoke godoc
// @Summary SSL Revoke Endpoint
// @Tags SSL
// @Accept json
// @Produce json
// @Router /ssl/revoke [post]
// @Param serial body string true "The Serial of the certificate to revoke"
// @Security API
// @Success 204
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) Revoke(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
