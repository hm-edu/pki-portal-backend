package ssl

import (
	"net/http"

	"github.com/hm-edu/portal-common/auth"

	"github.com/hm-edu/pki-rest-interface/pkg/model"
	pb "github.com/hm-edu/portal-apis"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
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

	domains, err := h.domain.ListDomains(c.Request().Context(), &pb.ListDomainsRequest{User: auth.UserFromRequest(c)})
	if err != nil {
		h.logger.Error("Error getting domains", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while listing certificates"}
	}
	_, err = h.ssl.ListCertificates(c.Request().Context(), &pb.ListSslRequest{Domains: domains.Domains})
	if err != nil {
		h.logger.Error("Error while listing certificates", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while listing certificates"}
	}
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
