package smime

import (
	"net/http"

	"github.com/hm-edu/pki-rest-interface/pkg/model"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/portal-common/logging"
	commonnModel "github.com/hm-edu/portal-common/model"
	"github.com/labstack/echo/v4"
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
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	ctx, span := h.tracer.Start(c.Request().Context(), "list")
	defer span.End()
	user := commonnModel.User{}
	if err := user.Bind(c, h.validator); err != nil {
		span.RecordError(err)
		return err
	}
	logger.Debug("Requesting smime certificates")
	certs, err := h.smime.ListCertificates(ctx, &pb.ListSmimeRequest{Email: user.Email})
	if err != nil {
		logger.Error("Error requesting smime certificates", zap.Error(err))
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
// @Param request body model.RevokeRequest true "The serial of the certificate to revoke and the reason"
// @Security API
// @Success 204
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) Revoke(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	ctx, span := h.tracer.Start(c.Request().Context(), "revoke")
	defer span.End()
	req := &model.RevokeRequest{}
	if err := req.Bind(c, h.validator); err != nil {
		span.RecordError(err)
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request"}
	}
	user := commonnModel.User{}
	if err := user.Bind(c, h.validator); err != nil {
		span.RecordError(err)
		return err
	}
	logger.Debug("Requesting smime certificate revocation")
	certs, err := h.smime.ListCertificates(ctx, &pb.ListSmimeRequest{Email: user.Email})

	if err != nil {
		span.RecordError(err)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Internal: err, Message: "Error processing the request"}
	}

	if !helper.Contains(helper.Map(certs.Certificates, func(t *pb.ListSmimeResponse_CertificateDetails) string { return t.Serial }), req.Serial) {
		logger.Warn("Certificate not found", zap.String("serial", req.Serial))
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request"}
	}

	_, err = h.smime.RevokeCertificate(ctx, &pb.RevokeSmimeRequest{Reason: req.Reason, Identifier: &pb.RevokeSmimeRequest_Serial{Serial: req.Serial}})
	if err != nil {
		logger.Error("Error requesting smime certificate revocation", zap.Error(err))
		span.RecordError(err)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Internal: err, Message: "Error processing the request"}
	}
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
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	ctx, span := h.tracer.Start(c.Request().Context(), "handleCsr")
	defer span.End()
	user := commonnModel.User{}
	if err := user.Bind(c, h.validator); err != nil {
		span.RecordError(err)
		return err
	}

	logger.Info("Requesting smime certificate")
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
		logger.Error("Error requesting smime certificate", zap.Error(err))
		span.RecordError(err)
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request"}
	}

	return c.JSON(http.StatusOK, cert.Certificate)
}
