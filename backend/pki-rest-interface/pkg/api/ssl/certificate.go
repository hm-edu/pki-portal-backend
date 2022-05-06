package ssl

import (
	"net/http"

	"github.com/hm-edu/pki-rest-interface/pkg/model"
	"github.com/hm-edu/portal-common/auth"
	"github.com/hm-edu/portal-common/helper"
	"go.opentelemetry.io/otel"

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

	domains, err := h.domain.ListDomains(c.Request().Context(), &pb.ListDomainsRequest{User: auth.UserFromRequest(c), Approved: true})
	if err != nil {
		h.logger.Error("Error getting domains", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while listing certificates"}
	}
	certs, err := h.ssl.ListCertificates(c.Request().Context(), &pb.ListSslRequest{Domains: domains.Domains})
	if err != nil {
		h.logger.Error("Error while listing certificates", zap.Error(err))
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
	ctx, span := otel.GetTracerProvider().Tracer("ssl").Start(c.Request().Context(), "revoke")
	defer span.End()
	req := &model.RevokeRequest{}
	if err := req.Bind(c, h.validator); err != nil {
		span.RecordError(err)
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid request"}
	}
	details, err := h.ssl.CertificateDetails(ctx, &pb.CertificateDetailsRequest{Serial: req.Serial})

	if err != nil {
		h.logger.Error("Error while revoking certificate", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while revoking certificate"}
	}

	domains, err := h.domain.ListDomains(ctx, &pb.ListDomainsRequest{User: auth.UserFromRequest(c), Approved: true})
	if err != nil {
		h.logger.Error("Error while revoking certificate", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while revoking certificate"}
	}

	for _, certDomain := range details.SubjectAlternativeNames {
		if !helper.Contains(domains.Domains, certDomain) {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "You are not authorized to revoke this certificate"}
		}
	}

	_, err = h.ssl.RevokeCertificate(ctx, &pb.RevokeSslRequest{Identifier: &pb.RevokeSslRequest_Serial{Serial: req.Serial}, Reason: req.Reason})
	if err != nil {
		h.logger.Error("Error while revoking certificate", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while revoking certificate"}
	}

	return c.NoContent(http.StatusNoContent)
}
