package model

import (
	"github.com/hm-edu/portal-common/model"
	"github.com/labstack/echo/v4"
)

// RevokeRequest holds the serial and the reason for revoking a SSL Certificate
type RevokeRequest struct {
	Serial string `json:"serial" validate:"required"`
	Reason string `json:"reason" validate:"required"`
}

// Bind binds an incoming echo request to the the CsrRequest and perfoms a validation
func (r *RevokeRequest) Bind(c echo.Context, v *model.Validator) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	err := v.Validate(r)
	return err
}
