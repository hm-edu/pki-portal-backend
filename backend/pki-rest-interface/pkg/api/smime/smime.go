package smime

import (
	"net/http"

	"github.com/hm-edu/pki-rest-interface/pkg/model"
	pb "github.com/hm-edu/portal-apis"
	commonnModel "github.com/hm-edu/portal-common/model"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

// List godoc
// @Summary SMIME List Endpoint
// @Tags SMIME
// @Accept json
// @Produce json
// @Router /smime/ [get]
// @Security API
// @Success 200 {object} []pb.ListSmimeResponse_CertificateDetails "certificates"
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) List(c echo.Context) error {
	ctx, span := otel.GetTracerProvider().Tracer("smime").Start(c.Request().Context(), "handleCsr")
	user := commonnModel.User{}
	if err := user.Bind(c, h.validator); err != nil {
		span.RecordError(err)
		return err
	}
	h.logger.Debug("Requesting smime certificates", zap.String("user", user.Email))
	certs, err := h.smime.ListCertificates(ctx, &pb.ListSmimeRequest{Email: user.Email})
	if err != nil {
		span.RecordError(err)
		return err
	}
	return c.JSON(http.StatusOK, certs.Certificates)
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
	ctx, span := otel.GetTracerProvider().Tracer("smime").Start(c.Request().Context(), "handleCsr")
	defer span.End()
	user := commonnModel.User{}
	if err := user.Bind(c, h.validator); err != nil {
		span.RecordError(err)
		return err
	}
	req := &model.CsrRequest{}
	if err := req.Bind(c, h.validator); err != nil {
		span.RecordError(err)
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request"}
	}
	cert, err := h.smime.IssueCertificate(ctx, &pb.IssueSmimeRequest{
		Csr:        req.CSR,
		Email:      user.Email,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		CommonName: user.CommonName,
	})
	if err != nil {
		span.RecordError(err)
		return err
	}

	return c.JSON(http.StatusOK, cert.Certificate)
}
