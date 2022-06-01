package domains

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/hm-edu/domain-rest-interface/ent"
	"github.com/hm-edu/domain-rest-interface/pkg/model"
	pb "github.com/hm-edu/portal-apis"
	"github.com/hm-edu/portal-common/auth"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/portal-common/logging"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// ListDomains godoc
// @Summary List domains.
// @Description Lists all domains that are either owned or delegated, or a child of a owned or delegated domain.
// @Tags Domains
// @Accept json
// @Produce json
// @Router /domains/ [get]
// @Security API
// @Success 200 {object} []model.Domain
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) ListDomains(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	ctx, span := h.tracer.Start(c.Request().Context(), "list")
	defer span.End()

	user, err := auth.UserFromRequest(c)
	if err != nil {
		logger.Error("Failed to get user from request", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid Request"}
	}
	span.SetAttributes(attribute.String("user", user))

	domains, err := h.enumerateDomains(ctx, user, logger)
	if err != nil {
		logger.Error("Listing domains failed", zap.Error(err))
		span.RecordError(err)
		return &echo.HTTPError{Internal: err, Code: http.StatusInternalServerError, Message: "Error while listing domains"}
	}
	logger.Debug("Listing domains", zap.Int("count", len(domains)), zap.Any("domains", domains))
	return c.JSON(http.StatusOK, domains)
}

func (h *Handler) enumerateDomains(ctx context.Context, user string, logger *zap.Logger) ([]*model.Domain, error) {

	ctx, span := h.tracer.Start(ctx, "enumerating")
	defer span.End()

	domains, err := h.domainStore.ListDomains(ctx, user, false, true)
	if err != nil {
		logger.Error("Listing domains failed", zap.Error(err))
		return nil, err
	}

	approvedDomains, err := h.domainStore.ListDomains(ctx, user, true, true)
	if err != nil {
		logger.Error("Listing domains failed", zap.Error(err))
		return nil, err
	}
	results := []*model.Domain{}

	filtered := helper.Where(approvedDomains, func(t *ent.Domain) bool { return t.Approved })

	certificates, err := h.pkiService.ListCertificates(ctx, &pb.ListSslRequest{IncludePartial: true, Domains: helper.Map(filtered, func(t *ent.Domain) string { return t.Fqdn })})
	if err != nil {
		logger.Error("Listing certificates failed", zap.Error(err))
		return nil, err
	}

	for _, domain := range domains {
		item := model.DomainToOutput(domain)
		// Users may always delete their own domains and transfer it.
		if item.Owner == user {
			item.Permissions.CanDelete = true
			if item.Approved {
				item.Permissions.CanTransfer = true
				item.Permissions.CanDelegate = true
			}
		}

		// Users may transfer or delete child domains
		if helper.Any(filtered, func(i *ent.Domain) bool { return strings.HasSuffix(domain.Fqdn, "."+i.Fqdn) }) {
			// Users may approve child domains
			if !item.Approved {
				item.Permissions.CanApprove = true
			}
			item.Permissions.CanDelete = true
			item.Permissions.CanTransfer = true
			item.Permissions.CanDelegate = true
		} else {
			// There is no upper domain for this user -> Prevent deletion
			if item.Permissions.CanDelete && item.Approved {
				item.Permissions.CanDelete = false
			}
		}
		if item.Permissions.CanDelete {
			matchingCerts := helper.Where(certificates.Items, func(t *pb.SslCertificateDetails) bool {
				return t.Status != "Revoked" && helper.Contains(t.SubjectAlternativeNames, item.FQDN)
			})
			if len(matchingCerts) > 0 {
				for _, cert := range matchingCerts {
					if cert.Id == 0 {
						logger.Debug("Certificate has no ID; Denying deletion of domain.", zap.String("serial", cert.Serial), zap.String("domain", domain.Fqdn), zap.String("user", user))
						item.Permissions.CanDelete = false
					}
					for _, name := range cert.SubjectAlternativeNames {
						if !helper.Any(filtered, func(i *ent.Domain) bool { return i.Fqdn == name }) {
							logger.Debug("Certificate has SAN without grant for user; Denying deletion of domain.", zap.String("serial", cert.Serial), zap.String("domain", domain.Fqdn), zap.String("san", name), zap.String("user", user))
							item.Permissions.CanDelete = false
							break
						}
					}
					if !item.Permissions.CanDelete {
						break
					}
				}
			}
		}
		results = append(results, &item)
	}
	return results, nil
}

// CreateDomain godoc
// @Summary Create a domain.
// @Description Creates a new domain if the FQDN is not already taken. Approvement is automatically done, in case the user owns a upper zone or a upper zone was already delegated to him.
// @Tags Domains
// @Accept json
// @Produce json
// @Router /domains/ [post]
// @Param domain body model.DomainRequest true "The Domain to create"
// @Security API
// @Success 201 {object} model.Domain
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) CreateDomain(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	ctx, span := h.tracer.Start(c.Request().Context(), "create")
	defer span.End()

	user, err := auth.UserFromRequest(c)
	if err != nil {
		logger.Error("Failed to get user from request", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid Request"}
	}

	span.SetAttributes(attribute.String("user", user))

	req := &model.DomainRequest{}
	if err := req.Bind(c, h.validator); err != nil {
		logger.Error("Binding request failed", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid Request"}
	}
	logger = logger.With(zap.String("fqdn", req.FQDN))

	allDomains, err := h.domainStore.ListAllDomains(ctx, false)
	if err != nil {
		logger.Error("Listing domains failed", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Internal: err, Message: "Failed to list domains"}
	}
	if helper.Contains(allDomains, req.FQDN) {
		logger.Error("Domain already exists", zap.String("fqdn", req.FQDN))
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Domain already exists"}
	}

	domain := ent.Domain{Owner: user, Fqdn: req.FQDN}

	domains, err := h.domainStore.ListDomains(ctx, user, true, true)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid Request"}
	}

	if helper.Any(domains, func(i *ent.Domain) bool { return i.Approved && strings.HasSuffix(domain.Fqdn, "."+i.Fqdn) }) {
		logger.Info("Auto approving domain request")
		domain.Approved = true
	}
	logger.Info("Creating domain", zap.Bool("approved", domain.Approved), zap.String("owner", domain.Owner))
	created, err := h.domainStore.Create(ctx, &domain)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest}
	}
	item := model.DomainToOutput(created)
	item.Permissions.CanDelete = true
	if item.Approved {
		item.Permissions.CanTransfer = true
		item.Permissions.CanDelegate = true
	}

	return c.JSON(http.StatusCreated, item)
}

// DeleteDomain godoc
// @Summary Delete a domain
// @Description Deletes a domain. Existing certificates are not are not longer accessible.
// @Tags Domains
// @Accept json
// @Produce json
// @Router /domains/{id} [delete]
// @Param id path int true "Domain ID"
// @Security API
// @Success 204
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) DeleteDomain(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	ctx, span := h.tracer.Start(c.Request().Context(), "delete")
	defer span.End()

	item, err := h.evaluatePermission(ctx, c, logger, func(d *model.Domain) bool { return d.Permissions.CanDelete })
	if err != nil {
		return err
	}

	logger.Info("Deleting domain", zap.String("fqdn", item.FQDN))
	if err := h.domainStore.Delete(ctx, item.ID); err != nil {
		logger.Error("Deleting domain failed", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Error while deleting domains"}
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) evaluatePermission(ctx context.Context, c echo.Context, logger *zap.Logger, predicate func(*model.Domain) bool) (*model.Domain, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Error("Invalid domain id", zap.Error(err))
		return nil, &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid domain ID"}
	}
	user, err := auth.UserFromRequest(c)
	if err != nil {
		logger.Error("Failed to get user from request", zap.Error(err))
		return nil, &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid Request"}
	}
	logger = logger.With(zap.Int("domain_id", id))
	domains, err := h.enumerateDomains(ctx, user, logger)
	if err != nil {
		return nil, &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error listing domains"}
	}

	item := helper.First(domains, func(i *model.Domain) bool { return i.ID == id })
	if item == nil {
		return nil, &echo.HTTPError{Code: http.StatusNotFound, Message: "Domain not found"}
	}
	if !predicate(item) {
		return nil, &echo.HTTPError{Code: http.StatusForbidden, Message: "Operation not allowed"}
	}
	return item, nil
}

// ApproveDomain godoc
// @Summary Approve domain request
// @Description Approves an outstanding domain request
// @Tags Domains
// @Accept json
// @Produce json
// @Router /domains/{id}/approve [post]
// @Param id path int true "Domain ID"
// @Security API
// @Success 200 {object} model.Domain The updated domain
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) ApproveDomain(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	ctx, span := h.tracer.Start(c.Request().Context(), "approve")
	defer span.End()

	item, err := h.evaluatePermission(ctx, c, logger, func(d *model.Domain) bool { return d.Permissions.CanApprove })
	if err != nil {
		return err
	}

	logger.Info("Approving domain", zap.String("fqdn", item.FQDN))
	updated, err := h.domainStore.Approve(ctx, item.ID)
	if err != nil {
		logger.Error("Approving domain failed", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Error while approving domain"}
	}

	return c.JSON(http.StatusOK, model.DomainToOutput(updated))
}

// TransferDomain godoc
// @Summary Transfer domain
// @Description Transfers a domain to a new owner. Transferring is only possible if you are either the owner of the domain itself or responsible for one of the parent zones.
// @Tags Domains
// @Accept json
// @Produce json
// @Router /domains/{id}/transfer [post]
// @Param domain body model.TransferRequest true "The Domain to transfer"
// @Param id path int true "Domain ID"
// @Security API
// @Success 200 {object} model.Domain The updated domain
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) TransferDomain(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	ctx, span := h.tracer.Start(c.Request().Context(), "transfer")
	defer span.End()

	req := &model.TransferRequest{}
	if err := req.Bind(c, h.validator); err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest}
	}

	item, err := h.evaluatePermission(ctx, c, logger, func(d *model.Domain) bool { return d.Permissions.CanTransfer })
	if err != nil {
		return err
	}
	logger.Info("Transferring domain", zap.String("fqdn", item.FQDN), zap.String("owner", req.Owner))

	updated, err := h.domainStore.Owner(ctx, item.ID, req.Owner)
	if err != nil {
		logger.Error("Transferring domain failed", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusBadRequest}
	}

	return c.JSON(http.StatusOK, model.DomainToOutput(updated))
}

// DeleteDelegation godoc
// @Summary Delete delegation.
// @Description Deletes an existing delegation.
// @Tags Domains
// @Accept json
// @Produce json
// @Router /domains/{id}/delegation/{delegation} [delete]
// @Param id path int true "Domain ID"
// @Param delegation path int true "Delegation ID"
// @Security API
// @Success 200 {object} model.Domain The updated domain
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) DeleteDelegation(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	ctx, span := h.tracer.Start(c.Request().Context(), "deleteDelegation")
	defer span.End()

	item, err := h.evaluatePermission(ctx, c, logger, func(d *model.Domain) bool { return d.Permissions.CanDelegate })
	if err != nil {
		return err
	}

	delegationID, err := strconv.Atoi(c.Param("delegation"))
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest}
	}
	logger.Info("Deleting delegation", zap.Int("domain", item.ID), zap.Int("delegation", delegationID), zap.Any("domain", item))

	delegation := helper.First(item.Delegations, func(t *model.Delegation) bool { return delegationID == t.ID })
	if delegation != nil {
		logger.Error("Delegation not found", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusNotFound, Message: "Delegation not found"}
	}
	logger.Info("Deleting delegation", zap.String("fqdn", item.FQDN), zap.Int("delegation", delegationID), zap.String("owner", delegation.User))
	updated, err := h.domainStore.DeleteDelegation(ctx, item.ID, delegationID)
	if err != nil {
		logger.Error("Deleting delegation failed", zap.Error(err))
		return &echo.HTTPError{Code: http.StatusNotFound, Message: "Deleting delegation failed"}
	}

	return c.JSON(http.StatusOK, model.DomainToOutput(updated))
}

// AddDelegation godoc
// @Summary Add delegation.
// @Description Adds a new delegation to an existing domain.
// @Tags Domains
// @Accept json
// @Produce json
// @Router /domains/{id}/delegation [post]
// @Param delegation body model.DelegationRequest true "The Delegation to add"
// @Param id path int true "Domain ID"
// @Security API
// @Success 200 {object} model.Domain The updated domain
// @Response default {object} echo.HTTPError "Error processing the request"
func (h *Handler) AddDelegation(c echo.Context) error {
	logger := c.Request().Context().Value(logging.LoggingContextKey).(*zap.Logger)
	ctx, span := h.tracer.Start(c.Request().Context(), "addDelegation")
	defer span.End()

	req := &model.DelegationRequest{}
	if err := req.Bind(c, h.validator); err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid Request"}
	}
	item, err := h.evaluatePermission(ctx, c, logger, func(d *model.Domain) bool { return d.Permissions.CanDelegate })
	if err != nil {
		return err
	}
	exists := helper.Any(item.Delegations, (func(t *model.Delegation) bool { return t.User == req.User }))
	if exists || item.Owner == req.User {
		logger.Error("Delegation already exists", zap.String("fqdn", item.FQDN), zap.String("owner", req.User))
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Delegation already exists"}
	}
	logger.Info("Adding delegation", zap.String("fqdn", item.FQDN), zap.String("owner", req.User))
	updated, err := h.domainStore.AddDelegation(ctx, item.ID, req.User)
	if err != nil {
		logger.Error("Adding delegation failed", zap.Error(err))

		return &echo.HTTPError{Code: http.StatusNotFound, Message: "Adding delegation failed"}
	}

	return c.JSON(http.StatusOK, model.DomainToOutput(updated))
}
