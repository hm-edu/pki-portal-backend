package api

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"

	"github.com/hm-edu/pki-service/pkg/model"
	commonnModel "github.com/hm-edu/portal-common/model"
	"github.com/hm-edu/sectigo-client/sectigo/client"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

func (h *Handler) HandleCsr(c echo.Context) error {
	user := commonnModel.User{}

	if err := user.Bind(c, h.validator); err != nil {
		return err
	}

	req := &model.CsrRequest{}
	if err := req.Bind(c, h.validator); err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request"}
	}

	block, _ := pem.Decode([]byte(req.CSR))

	// Validate the passed CSR to comply the server-side requirements (e.g. key-strength, key-type, etc.)
	// The "real" user-data will be filled in by sectigo so we can sort of ignore any data provided by the user and simply pass the CSR to sectigo
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid CSR"}
	}

	// Validate the CSR
	if err := csr.CheckSignature(); err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid CSR"}
	}

	if csr.PublicKeyAlgorithm != x509.RSA {
		return echo.NewHTTPError(http.StatusBadRequest, "Only RSA keys are supported")
	}

	// Get the public key from the CSR
	pubKey, ok := csr.PublicKey.(*rsa.PublicKey)
	size := pubKey.Size() * 8
	if !ok || size != 4096 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid CSR")
	}

	resp, err := h.sectigoClient.ClientService.Enroll(client.EnrollmentRequest{
		OrgID:           viper.GetInt("smime_org_id"),
		FirstName:       user.Firstname,
		MiddleName:      "",
		CommonName:      user.CommonName,
		LastName:        user.Lastname,
		Email:           user.Email,
		Phone:           "",
		SecondaryEmails: []string{},
		CSR:             req.CSR,
		CertType:        viper.GetInt("smime_cert_type"),
		Term:            viper.GetInt("smime_term"),
		Eppn:            "",
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	certificate, err := h.sectigoClient.ClientService.Collect(resp.OrderNumber, "x509")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, certificate)
}
