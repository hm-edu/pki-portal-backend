package model

import (
	"github.com/hm-edu/domain-rest-interface/ent"
	"github.com/hm-edu/portal-common/helper"
	"github.com/hm-edu/portal-common/model"
	"github.com/labstack/echo/v4"
)

// DomainRequest represents an request for an action on a domain.
type DomainRequest struct {
	FQDN string `json:"fqdn" validate:"required,fqdn"`
}

// Bind binds an incoming echo request to the internal domain model and perfoms a validation
func (r *DomainRequest) Bind(c echo.Context, v *model.Validator) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	err := v.Validate(r)
	return err
}

// DomainToOutput converts the internal domain model to the REST representation.
func DomainToOutput(d *ent.Domain) Domain {
	return Domain{ID: d.ID, FQDN: d.Fqdn, Owner: d.Owner, Approved: d.Approved, Delegations: helper.Map(d.Edges.Delegations, DelegationToOutput)}
}

// Domain represents a domain.
type Domain struct {
	ID          int          `json:"id"`
	FQDN        string       `json:"fqdn"`
	Owner       string       `json:"owner"`
	Delegations []Delegation `json:"delegations"`
	Approved    bool         `json:"approved"`
}
