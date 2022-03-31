package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hm-edu/domain-service/ent"
	"github.com/hm-edu/domain-service/pkg/model"
	"github.com/hm-edu/portal-common/auth"
	"github.com/hm-edu/portal-common/helper"
	"github.com/labstack/echo/v4"
)

// ListDomains godoc
// @Summary List domains endpoint
// @Description Lists all domains
// @Tags Domains
// @Accept json
// @Produce json
// @Router /domains/ [get]
// @Security API
// @Success 200 {object} []model.Domain
// @Failure 400 {object} echo.HTTPError "Unauthorized"
func (h *Handler) ListDomains(c echo.Context) error {
	domains, err := h.domainStore.ListDomains(c.Request().Context(), auth.UserFromRequest(c))
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest}
	}
	return c.JSON(http.StatusOK, helper.Map(domains, model.DomainToOutput))
}

// CreateDomain godoc
// @Summary Creates a new domain if the FQDN is not already taken
// @Description Creates a new domain
// @Tags Domains
// @Accept json
// @Produce json
// @Router /domains/ [post]
// @Param domain body model.DomainRequest true "The Domain to create"
// @Security API
// @Success 201 {object} model.Domain
// @Failure 400 {object} echo.HTTPError "Unauthorized"
func (h *Handler) CreateDomain(c echo.Context) error {
	var new ent.Domain
	req := &model.DomainRequest{}
	if err := req.Bind(c, &new, h.validator); err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid Request"}
	}
	new.Owner = auth.UserFromRequest(c)

	domains, err := h.domainStore.ListDomains(c.Request().Context(), auth.UserFromRequest(c))
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest}
	}

	if helper.Any(domains, func(i *ent.Domain) bool { return i.Approved && strings.HasSuffix(new.Fqdn, "."+i.Fqdn) }) {
		new.Approved = true
	}

	err = h.domainStore.CreateDomain(c.Request().Context(), &new)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest}
	}
	return c.JSON(http.StatusCreated, model.DomainToOutput(&new))
}

// DeleteDomains godoc
// @Summary Delete a domain and optional the complete subtree
// @Description Deletes a domain
// @Tags Domains
// @Accept json
// @Produce json
// @Router /domains/ [delete]
// @Param domain body model.DeleteDomainRequest true "The Domains to delete"
// @Security API
// @Success 204
// @Failure 400 {object} echo.HTTPError "Unauthorized or Request Error"
func (h *Handler) DeleteDomains(c echo.Context) error {
	req := &model.DeleteDomainRequest{}
	if err := req.Bind(c, h.validator); err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid Request"}
	}
	domains, err := h.domainStore.ListDomains(c.Request().Context(), auth.UserFromRequest(c))
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest}
	}
	d := helper.Where(domains, func(i *ent.Domain) bool {
		if req.SubTree {
			return strings.HasSuffix(i.Fqdn, fmt.Sprintf(".%s", req.FQDN)) || i.Fqdn == req.FQDN
		} else {
			return i.Fqdn == req.FQDN
		}
	})
	if len(d) != 0 {
		if err := h.domainStore.DeleteDomains(c.Request().Context(), d); err != nil {
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Error while deleting domains"}
		}
		return c.NoContent(http.StatusNoContent)
	}
	return &echo.HTTPError{Code: http.StatusBadRequest, Message: "No domains found"}
}

// ApproveDomain godoc
// @Summary Approves a domain request
// @Description Approves an outstanding domain request
// @Tags Domains
// @Accept json
// @Produce json
// @Router /domains/approve [post]
// @Param domain body model.DomainRequest true "The Domain to approve"
// @Security API
// @Success 200 {object} model.Domain The updated domain
// @Failure 400 {object} echo.HTTPError "Unauthorized or Bad Request"
// @Failure 403 {object} echo.HTTPError "Access to domain denied"
// @Failure 404 {object} echo.HTTPError "Domain in zone does not exist"
func (h *Handler) ApproveDomain(c echo.Context) error {
	var new ent.Domain
	req := &model.DomainRequest{}
	if err := req.Bind(c, &new, h.validator); err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Internal: err, Message: "Invalid Request"}
	}

	domains, err := h.domainStore.ListDomains(c.Request().Context(), auth.UserFromRequest(c))
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest}
	}

	if !helper.Any(domains, func(i *ent.Domain) bool { return i.Approved && strings.HasSuffix(new.Fqdn, "."+i.Fqdn) }) {
		return &echo.HTTPError{Code: http.StatusForbidden, Message: "Your not responsible for this zone"}
	}

	domain, err := h.domainStore.GetDomain(c.Request().Context(), req.FQDN)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusNotFound, Message: "Your not responsible for this zone"}
	}

	if err := h.domainStore.Approve(c.Request().Context(), domain); err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest}
	}

	return c.JSON(http.StatusOK, model.DomainToOutput(domain))
}

// TransferDomain godoc
// @Summary Transfers a domain to a new owner
// @Description Transfers a domain to a new owner
// @Tags Domains
// @Accept json
// @Produce json
// @Router /domains/transfer [post]
// @Param domain body model.DomainRequest true "The Domain to transfer"
// @Security API
// @Success 204
// @Failure 400 {object} echo.HTTPError "Unauthorized"
func (h *Handler) TransferDomain(c echo.Context) error {
	//domains, err := h.domainStore.ListDomains(c.Request().Context(), auth.UserFromRequest(c))
	//if err != nil {
	//	return &echo.HTTPError{Code: http.StatusBadRequest}
	//}
	return c.NoContent(http.StatusAccepted)
}
