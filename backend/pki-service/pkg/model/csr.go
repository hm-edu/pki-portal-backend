package model

import (
	"github.com/hm-edu/portal-common/model"
	"github.com/labstack/echo/v4"
)

// CsrRequest holds a CSR.
type CsrRequest struct {
	CSR string `json:"csr" validate:"required"`
}

// Bind binds an incoming echo request to the the CsrRequest and perfoms a validation
func (r *CsrRequest) Bind(c echo.Context, v *model.Validator) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	err := v.Validate(r)
	return err
}
