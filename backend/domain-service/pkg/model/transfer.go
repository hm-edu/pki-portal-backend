package model

import "github.com/labstack/echo/v4"

// TransferRequest represents an request for transfering a domain.
type TransferRequest struct {
	Owner string `json:"owner" validate:"required"`
}

// Bind binds an incoming echo request to the the TransferRequest and perfoms a validation
func (r *TransferRequest) Bind(c echo.Context, v *Validator) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	err := v.Validate(r)
	return err
}
