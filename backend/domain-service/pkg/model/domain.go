package model

import (
	"github.com/go-playground/validator/v10"
	"github.com/hm-edu/domain-service/ent"
	"github.com/hm-edu/portal-common/helper"
	"github.com/labstack/echo/v4"
)

type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{
		validator: validator.New(),
	}
}

func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

type DomainRequest struct {
	FQDN string `json:"fqdn" validate:"required,fqdn"`
}
type DeleteDomainRequest struct {
	FQDN    string `json:"fqdn" validate:"required,fqdn"`
	SubTree bool   `json:"subtree"`
}

func (r *DomainRequest) Bind(c echo.Context, d *ent.Domain, v *Validator) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := v.Validate(r); err != nil {
		return err
	}
	d.Fqdn = r.FQDN
	return nil
}

func (r *DeleteDomainRequest) Bind(c echo.Context, v *Validator) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := v.Validate(r); err != nil {
		return err
	}
	return nil
}

func DelegationToOutput(d *ent.Delegation) Delegation {
	return Delegation{ID: d.ID, User: d.User}
}

func DomainToOutput(d *ent.Domain) Domain {
	return Domain{ID: d.ID, FQDN: d.Fqdn, Owner: d.Owner, Approved: d.Approved, Delegations: helper.Map(d.Edges.Delegations, DelegationToOutput)}
}

type Delegation struct {
	ID   int    `json:"id"`
	User string `json:"user"`
}

type Domain struct {
	ID          int          `json:"id"`
	FQDN        string       `json:"fqdn"`
	Owner       string       `json:"owner"`
	Delegations []Delegation `json:"delegations"`
	Approved    bool         `json:"approved"`
}
