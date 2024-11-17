package ssl

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"

	sentryecho "github.com/getsentry/sentry-go/echo"

	"github.com/getsentry/sentry-go"
	"github.com/hm-edu/pki-rest-interface/pkg/model"
	"github.com/hm-edu/portal-common/auth"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/portal-common/logging"

	pb "github.com/hm-edu/portal-apis"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Active godoc
// @Summary SSL List active certificates Endpoint
// @Tags SSL
// @Accept json
// @Produce json
// @Router /ssl/active [get]
// @Param        domain    query     string  true  "domain search by domain"
// @Security API
// @Success 200 {object} []pb.SslCertificateDetails "Certificates"
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) Active(c echo.Context) error {

	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	user, err := auth.UserFromRequest(c)
	if err != nil {
		logger.Error("error getting user from request", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid Request"}
	}

	if hub := sentryecho.GetHubFromContext(c); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetExtra("user", user)
		})
	}
	span := sentryecho.GetSpanFromContext(c)
	ctx := span.Context()
	domain := c.QueryParam("domain")
	if domain == "" {
		logger.Warn("no domain provided")
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "No domain provided"}
	}

	sentry.AddBreadcrumb(&sentry.Breadcrumb{Level: sentry.LevelInfo, Message: "Loading domains"})
	domains, err := h.domain.ListDomains(ctx, &pb.ListDomainsRequest{User: user, Approved: true})
	if err != nil {
		logger.Error("error getting domains", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while listing certificates"}
	}

	if !helper.Contains(domains.Domains, domain) {
		logger.Warn("domain not found. Usage not allowed.", zap.String("domain", domain))
		return &echo.HTTPError{Code: http.StatusForbidden, Message: "You are not authorized to use this domain"}
	}

	sentry.AddBreadcrumb(&sentry.Breadcrumb{Level: sentry.LevelInfo, Message: "Loading certificates"})

	logger.Info("fetching certificates", zap.Strings("domains", domains.Domains))
	certs, err := h.ssl.ListCertificates(ctx, &pb.ListSslRequest{IncludePartial: false, Domains: []string{domain}})
	if err != nil {
		logger.Error("error while listing certificates", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while listing certificates"}
	}

	active := make([]*pb.SslCertificateDetails, 0)
	for _, cert := range certs.Items {
		if cert.Status == "Issued" {
			active = append(active, cert)
		}
	}
	return c.JSON(http.StatusOK, active)
}

// List godoc
// @Summary SSL List Endpoint
// @Tags SSL
// @Accept json
// @Produce json
// @Router /ssl/ [get]
// @Security API
// @Success 200 {object} []pb.SslCertificateDetails "Certificates"
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) List(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)

	span := sentryecho.GetSpanFromContext(c)
	user, err := auth.UserFromRequest(c)
	if err != nil {
		logger.Error("error getting user from request", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid Request"}
	}
	if hub := sentryecho.GetHubFromContext(c); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetExtra("user", user)
		})
	}
	sentry.AddBreadcrumb(&sentry.Breadcrumb{Level: sentry.LevelInfo, Message: "Loading domains"})
	domains, err := h.domain.ListDomains(span.Context(), &pb.ListDomainsRequest{User: user, Approved: true})
	if err != nil {
		logger.Error("error getting domains", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while listing certificates"}
	}

	logger.Debug("fetching certificates", zap.Strings("domains", domains.Domains))
	sentry.AddBreadcrumb(&sentry.Breadcrumb{Level: sentry.LevelInfo, Message: "Loading certificates"})
	certs, err := h.ssl.ListCertificates(span.Context(), &pb.ListSslRequest{IncludePartial: false, Domains: domains.Domains})
	if err != nil {
		logger.Error("error while listing certificates", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while listing certificates"}
	}

	return c.JSON(http.StatusOK, certs.Items)
}

// Revoke godoc
// @Summary SSL Revoke Endpoint
// @Tags SSL
// @Accept json
// @Produce json
// @Router /ssl/revoke [post]
// @Param request body model.RevokeRequest true "The serial of the certificate to revoke and the reason"
// @Security API
// @Success 204
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) Revoke(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	user, err := auth.UserFromRequest(c)
	if err != nil {
		logger.Error("error getting user from request", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid Request"}
	}
	if hub := sentryecho.GetHubFromContext(c); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetExtra("user", user)
		})
	}

	span := sentryecho.GetSpanFromContext(c)
	ctx := span.Context()

	req := &model.RevokeRequest{}
	if err := req.Bind(c, h.validator); err != nil {
		logger.Error("error while validating request", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request"}
	}

	logger.Info("trying to revoke certificate", zap.String("serial", req.Serial), zap.String("reason", req.Reason))

	details, err := h.ssl.CertificateDetails(ctx, &pb.CertificateDetailsRequest{Serial: req.Serial})

	if err != nil {
		sentry.CaptureException(err)
		logger.Error("error while revoking certificate", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while revoking certificate"}
	}
	domains, err := h.domain.ListDomains(ctx, &pb.ListDomainsRequest{User: user, Approved: true})
	if err != nil {
		sentry.CaptureException(err)
		logger.Error("error listing domains for certificate revocation", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while revoking certificate"}
	}

	for _, certDomain := range details.SubjectAlternativeNames {
		if !helper.Contains(domains.Domains, certDomain) {
			logger.Warn("domain not found. Revocation not allowed.", zap.String("domain", certDomain))
			return &echo.HTTPError{Code: http.StatusForbidden, Message: "You are not authorized to revoke this certificate"}
		}
	}
	_, err = h.ssl.RevokeCertificate(ctx, &pb.RevokeSslRequest{Identifier: &pb.RevokeSslRequest_Serial{Serial: req.Serial}, Reason: req.Reason})
	if err != nil {
		sentry.CaptureException(err)
		logger.Error("error while revoking certificate", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while revoking certificate"}
	}

	logger.Info("certificate revoked", zap.String("serial", req.Serial))
	return c.NoContent(http.StatusNoContent)
}

// HandleCsr godoc
// @Summary SSL CSR Endpoint
// @Description This endpoint handles a provided CSR. The validity of the CSR is checked and passed to the sectigo server.
// @Tags SSL
// @Accept json
// @Produce json
// @Router /ssl/csr [post]
// @Param csr body model.CsrRequest true "The CSR"
// @Security API
// @Success 200 {string} string "certificate"
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) HandleCsr(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	user, err := auth.UserFromRequest(c)
	if err != nil {
		sentry.CaptureException(err)
		logger.Error("error getting user from request", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid Request"}
	}

	if hub := sentryecho.GetHubFromContext(c); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetExtra("user", user)
		})
	}

	span := sentryecho.GetSpanFromContext(c)
	ctx := span.Context()

	req := &model.CsrRequest{}
	if err := req.Bind(c, h.validator); err != nil {
		sentry.CaptureException(err)
		logger.Error("error while parsing csr", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request"}
	}
	block, _ := pem.Decode([]byte(req.CSR))

	if block == nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request. CSR has invalid format."}
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		logger.Error("error while parsing csr", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request. CSR has invalid format."}
	}

	// Validate the CSR
	if err := csr.CheckSignature(); err != nil {
		sentry.CaptureException(err)
		logger.Error("error while parsing csr", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request. CSR has invalid signature."}
	}

	// Get the public key from the CSR
	pubKey, ok := csr.PublicKey.(*rsa.PublicKey)
	if ok {
		size := pubKey.Size() * 8
		if ok && size < 2048 {
			logger.Warn("Invalid key length", zap.String("key_length", fmt.Sprintf("%d", size)))
			return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request. RSA key length must be greater than 2048."}
		}
	}
	sans := make([]string, 0, len(csr.DNSNames)+len(csr.IPAddresses)+len(csr.URIs)+1)
	if csr.Subject.CommonName != "" {
		sans = append(sans, csr.Subject.CommonName)
	}
	sans = append(sans, csr.DNSNames...)
	for _, ip := range csr.IPAddresses {
		sans = append(sans, ip.String())
	}
	for _, u := range csr.URIs {
		sans = append(sans, u.String())
	}
	if len(sans) == 0 {
		logger.Info("no SANs found in CSR")
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request. No SANs found in CSR"}
	}

	permissions, err := h.domain.CheckPermission(ctx, &pb.CheckPermissionRequest{User: user, Domains: sans})
	if err != nil {
		sentry.CaptureException(err)
		logger.Error("error while checking permissions", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while checking permissions"}
	}
	logger.Info("checking permission for certificate issuance", zap.Strings("domains", sans))
	missing := helper.Map(helper.Where(permissions.Permissions, func(t *pb.Permission) bool { return !t.Granted }), func(t *pb.Permission) string { return t.Domain })
	logger.Info("permissions checked", zap.Strings("missing", missing), zap.Strings("domains", sans))
	if len(missing) > 0 {
		return &echo.HTTPError{Code: http.StatusForbidden, Message: "You are not authorized to issue this certificate. Missing permissions for domains: " + strings.Join(missing, ", ")}
	}

	resp, err := h.ssl.IssueCertificate(ctx, &pb.IssueSslRequest{Csr: req.CSR, SubjectAlternativeNames: sans, Issuer: user, Source: "API"})
	if err != nil {
		sentry.CaptureException(err)
		logger.Error("error while processing CSR", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Internal Error while processing the request."}
	}
	return c.JSON(http.StatusOK, resp.Certificate)
}
