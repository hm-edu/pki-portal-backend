package model

import (
	"github.com/hm-edu/domain-service/ent"
	"github.com/labstack/echo/v4"
)

// DelegationRequest represents an request for either adding or removing a delegation from a domain.
type DelegationRequest struct {
	User string `json:"user" validate:"required"`
}

// Bind binds an incoming echo request to the the TransferRequest and perfoms a validation
func (r *DelegationRequest) Bind(c echo.Context, v *Validator) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	err := v.Validate(r)
	return err
}

// Delegation represents a delegation of a domain to a user.
type Delegation struct {
	ID   int    `json:"id"`
	User string `json:"user"`
}

// DelegationToOutput converts the internal delegation model to the REST representation.
func DelegationToOutput(d *ent.Delegation) Delegation {
	return Delegation{ID: d.ID, User: d.User}
}
