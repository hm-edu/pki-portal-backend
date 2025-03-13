package smime

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"net/http"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/hm-edu/pki-rest-interface/pkg/model"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/portal-common/logging"
	commonModel "github.com/hm-edu/portal-common/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// List godoc
// @Summary SMIME List Endpoint
// @Tags SMIME
// @Accept json
// @Produce json
// @Router /smime/ [get]
// @Param request query model.ListSmimeCertificatesRequest true "The email to list smime certificates for"
// @Security API
// @Success 200 {object} []pb.ListSmimeResponse_CertificateDetails "certificates"
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) List(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)

	req := &model.ListSmimeCertificatesRequest{}
	hub := sentryecho.GetHubFromContext(c)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	user := commonModel.User{}
	if err := user.Bind(c, h.validator); err != nil {
		hub.CaptureException(err)
		return err
	}

	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{Email: user.Email})
	})
	if err := req.Bind(c, h.validator); err != nil {
		hub.CaptureException(err)
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request"}
	}

	span := sentryecho.GetSpanFromContext(c)
	ctx := c.Request().Context()
	if span != nil {
		ctx = span.Context()
	}
	// use requested email if provided, otherwise use the user's primary email
	requestedEmail := user.Email
	if len(req.Email) > 0 {
		requestedEmail = req.Email
	}

	// check if the requested email belongs to the user
	if requestedEmail != user.Email && !helper.Contains(user.AdditionalSmimeEmails, requestedEmail) {
		return &echo.HTTPError{Code: http.StatusForbidden, Message: "You are not authorized to use this email."}
	}

	logger.Debug("requesting smime certificates")
	certs, err := h.smime.ListCertificates(ctx, &pb.ListSmimeRequest{Email: requestedEmail})
	if err != nil {
		hub.CaptureException(err)
		logger.Error("error requesting smime certificates", zap.Error(err))
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

	hub := sentryecho.GetHubFromContext(c)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}

	user := commonModel.User{}
	if err := user.Bind(c, h.validator); err != nil {
		hub.CaptureException(err)
		return err
	}

	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{Email: user.Email})
	})

	req := &model.RevokeRequest{}
	if err := req.Bind(c, h.validator); err != nil {
		hub.CaptureException(err)
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request"}
	}

	span := sentryecho.GetSpanFromContext(c)
	ctx := c.Request().Context()
	if span != nil {
		ctx = span.Context()
	}

	logger.Debug("requesting smime certificate revocation")
	listRes, err := h.smime.ListCertificates(ctx, &pb.ListSmimeRequest{Email: user.Email})
	if err != nil {
		hub.CaptureException(err)
		return &echo.HTTPError{Code: http.StatusInternalServerError, Internal: err, Message: "Error processing the request"}
	}

	// list certificates for all emails
	certs := listRes.Certificates
	for _, email := range user.AdditionalSmimeEmails {
		listRes, err := h.smime.ListCertificates(ctx, &pb.ListSmimeRequest{Email: email})
		if err != nil {
			hub.CaptureException(err)
			continue
		}

		certs = append(certs, listRes.Certificates...)
	}

	if !helper.Contains(helper.Map(certs, func(t *pb.ListSmimeResponse_CertificateDetails) string { return t.Serial }), req.Serial) {
		logger.Warn("certificate not found", zap.String("serial", req.Serial))
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request"}
	}

	_, err = h.smime.RevokeCertificate(ctx, &pb.RevokeSmimeRequest{Reason: req.Reason, Identifier: &pb.RevokeSmimeRequest_Serial{Serial: req.Serial}})
	if err != nil {
		hub.CaptureException(err)
		logger.Error("error requesting smime certificate revocation", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Internal: err, Message: "Error processing the request"}
	}
	return c.NoContent(http.StatusNoContent)
}

// HandleCsr godoc
// @Summary SMIME CSR Endpoint
// @Description This endpoint handles a provided CSR. The validity of the CSR is checked and passed to the harica server in combination with the basic user information extracted from the JWT.
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
	hub := sentryecho.GetHubFromContext(c)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	user := commonModel.User{}
	if err := user.Bind(c, h.validator); err != nil {
		hub.CaptureException(err)
		return err
	}

	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{Email: user.Email})
	})

	span := sentryecho.GetSpanFromContext(c)
	ctx := c.Request().Context()
	if span != nil {
		ctx = span.Context()
	}

	if user.Student && h.rejectStudents {
		logger.Warn("rejecting student", zap.String("email", user.Email))
		return &echo.HTTPError{Code: http.StatusForbidden, Message: "Students are not allowed to request smime certificates"}
	}

	req := &model.CsrRequest{}
	if err := req.Bind(c, h.validator); err != nil {
		hub.CaptureException(err)
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request"}
	}

	block, _ := pem.Decode([]byte(req.CSR))
	if block == nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid request. CSR has invalid format."}
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		hub.CaptureException(err)
		logger.Error("error while parsing csr", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request. CSR has invalid format."}
	}

	// Validate the CSR
	if err := csr.CheckSignature(); err != nil {
		hub.CaptureException(err)
		logger.Error("error while parsing csr", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request. CSR has invalid signature."}
	}

	// check if the requested email belongs to the user
	requestedEmail := user.Email
	emailSubjects := helper.Where(csr.Subject.Names, func(a pkix.AttributeTypeAndValue) bool { return a.Type.String() == "1.2.840.113549.1.9.1" })
	if len(emailSubjects) > 0 {
		requestedEmail = emailSubjects[0].Value.(string)
	}
	if requestedEmail != user.Email && !helper.Contains(user.AdditionalSmimeEmails, requestedEmail) {
		return &echo.HTTPError{Code: http.StatusForbidden, Message: "You are not authorized to use this email."}
	}

	logger.Info("Issuing new smime certificate")
	cert, err := h.smime.IssueCertificate(ctx, &pb.IssueSmimeRequest{
		Csr:        req.CSR,
		Email:      requestedEmail,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
		CommonName: user.CommonName,
		Student:    user.Student,
	})
	if err != nil {
		hub.CaptureException(err)
		logger.Error("error requesting smime certificate", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Internal: err, Message: "Handling CSR failed"}
	}

	return c.JSON(http.StatusOK, cert.Certificate)
}
