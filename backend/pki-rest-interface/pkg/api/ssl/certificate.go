package ssl

import (
	"net/http"

	"github.com/hm-edu/pki-rest-interface/pkg/model"
	"github.com/hm-edu/portal-common/auth"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/portal-common/logging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

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
// @Success 200 {object} []pb.SslCertificateDetails "Certificates"
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) List(c echo.Context) error {
	logger := h.logger.With(logging.AddMetadata(c)...)
	ctx, span := h.tracer.Start(c.Request().Context(), "list")
	defer span.End()
	user, err := auth.UserFromRequest(c)
	if err != nil {
		h.logger.Error("Error getting user from request", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid Request"}
	}
	span.SetAttributes(attribute.String("user", user))
	logger = logger.With(zap.String("user", user))
	span.AddEvent("fetching domains")
	domains, err := h.domain.ListDomains(ctx, &pb.ListDomainsRequest{User: user, Approved: true})
	if err != nil {
		logger.Error("Error getting domains", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while listing certificates"}
	}
	span.AddEvent("fetching certificates")
	certs, err := h.ssl.ListCertificates(ctx, &pb.ListSslRequest{IncludePartial: false, Domains: domains.Domains})
	if err != nil {
		logger.Error("Error while listing certificates", zap.Error(err))
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
	logger := h.logger.With(logging.AddMetadata(c)...)
	ctx, span := h.tracer.Start(c.Request().Context(), "revoke")
	defer span.End()
	user, err := auth.UserFromRequest(c)
	if err != nil {
		h.logger.Error("Error getting user from request", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid Request"}
	}
	span.SetAttributes(attribute.String("user", user))
	logger = logger.With(zap.String("user", user))

	req := &model.RevokeRequest{}
	span.AddEvent("validating request")
	if err := req.Bind(c, h.validator); err != nil {
		span.RecordError(err)
		h.logger.Error("Error while validating request", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request"}
	}

	logger.Info("trying to revoke certificate", zap.String("serial", req.Serial), zap.String("reason", req.Reason))

	span.AddEvent("obtaining certificate details")
	details, err := h.ssl.CertificateDetails(ctx, &pb.CertificateDetailsRequest{Serial: req.Serial})

	if err != nil {
		span.RecordError(err)
		logger.Error("Error while revoking certificate", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while revoking certificate"}
	}
	span.AddEvent("fetching domains")
	domains, err := h.domain.ListDomains(ctx, &pb.ListDomainsRequest{User: user, Approved: true})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error listing domains")
		logger.Error("Error listing domains for certificate revocation", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while revoking certificate"}
	}

	for _, certDomain := range details.SubjectAlternativeNames {
		if !helper.Contains(domains.Domains, certDomain) {
			logger.Warn("Domain not found. Revocation not allowed.", zap.String("domain", certDomain))
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "You are not authorized to revoke this certificate"}
		}
	}
	span.AddEvent("revoking certificate")
	_, err = h.ssl.RevokeCertificate(ctx, &pb.RevokeSslRequest{Identifier: &pb.RevokeSslRequest_Serial{Serial: req.Serial}, Reason: req.Reason})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error revoking certificate")
		logger.Error("Error while revoking certificate", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while revoking certificate"}
	}

	return c.NoContent(http.StatusNoContent)
}
