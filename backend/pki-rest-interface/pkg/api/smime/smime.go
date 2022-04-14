package smime

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"

	"github.com/hm-edu/pki-rest-interface/pkg/model"
	commonnModel "github.com/hm-edu/portal-common/model"
	"github.com/hm-edu/sectigo-client/sectigo/client"
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
	_, span := otel.GetTracerProvider().Tracer("smime").Start(c.Request().Context(), "operation")
	defer span.End()
	span.AddEvent("Validation")
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

	block, _ := pem.Decode([]byte(req.CSR))

	// Validate the passed CSR to comply the server-side requirements (e.g. key-strength, key-type, etc.)
	// The "real" user-data will be filled in by sectigo so we can sort of ignore any data provided by the user and simply pass the CSR to sectigo
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		span.RecordError(err)
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid CSR"}
	}

	// Validate the CSR
	if err := csr.CheckSignature(); err != nil {
		span.RecordError(err)
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid CSR"}
	}

	if csr.PublicKeyAlgorithm != x509.RSA {
		return echo.NewHTTPError(http.StatusBadRequest, "Only RSA keys are supported")
	}

	// Get the public key from the CSR
	pubKey, ok := csr.PublicKey.(*rsa.PublicKey)
	size := pubKey.Size() * 8
	if !ok || fmt.Sprintf("%d", size) != h.sectigoCfg.SmimeKeyLength {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid CSR")
	}

	span.AddEvent("SectigoEnroll")
	resp, err := h.sectigoClient.ClientService.Enroll(client.EnrollmentRequest{
		OrgID:           h.sectigoCfg.SmimeOrgID,
		FirstName:       user.Firstname,
		MiddleName:      "",
		CommonName:      user.CommonName,
		LastName:        user.Lastname,
		Email:           user.Email,
		Phone:           "",
		SecondaryEmails: []string{},
		CSR:             req.CSR,
		CertType:        h.sectigoCfg.SmimeProfile,
		Term:            h.sectigoCfg.SmimeTerm,
		Eppn:            "",
	})
	if err != nil {
		span.RecordError(err)
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	span.AddEvent("SectigoCollect")
	certificate, err := h.sectigoClient.ClientService.Collect(resp.OrderNumber, "x509")
	if err != nil {
		span.RecordError(err)
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
	}

	return c.JSON(http.StatusOK, certificate)
}
