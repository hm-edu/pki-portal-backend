package smime

import (
	"net/http"

	"github.com/hm-edu/pki-rest-interface/pkg/model"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
)

// List godoc
// @Summary SMIME List Endpoint
// @Tags SMIME
// @Accept json
// @Produce json
// @Router /smime/ [get]
// @Security API
// @Success 200 {object} model.SMIME "certificate"
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) List(c echo.Context) error {
	return c.JSON(http.StatusOK, model.SMIME{})
}

// Revoke godoc
// @Summary SMIME Revoke Endpoint
// @Tags SMIME
// @Accept json
// @Produce json
// @Router /smime/revoke [post]
// @Param serial body string true "The serial of the certificate to revoke"
// @Security API
// @Success 204
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) Revoke(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// HandleCsr godoc
// @Summary SMIME CSR Endpoint
// @Description This endpoint handles a provided CSR. The validity of the CSR is checked and passed to the sectigo server in combination with the basic user information extracted from the JWT.
// @Description The server uses his own configuration, so the profile and the lifetime of the certificate can not be modified.
// @Description Afterwards the new certificate is returned as X509 certificate.
// @Tags SMIME
// @Accept json
// @Produce json
// @Router /smime/csr [post]
// @Param csr body model.CsrRequest true "The CSR"
// @Security API
// @Success 200 {string} string "certificate"
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) HandleCsr(c echo.Context) error {
	_, span := otel.GetTracerProvider().Tracer("smime").Start(c.Request().Context(), "handleCsr")
	defer span.End()
	return c.NoContent(http.StatusNoContent)
}
