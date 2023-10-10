package model

import (
	"github.com/hm-edu/portal-common/model"
	"github.com/labstack/echo/v4"
)

// ListSmimeCertificatesRequest represents an request for listing smime certificates.
type ListSmimeCertificatesRequest struct {
	Email string `query:"email"`
}

// Bind binds an incoming echo request to the ListSmimeCertificatesRequest and perfoms a validation
func (r *ListSmimeCertificatesRequest) Bind(c echo.Context, v *model.Validator) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	err := v.Validate(r)
	return err
}
