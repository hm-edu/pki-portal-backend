package models

import (
	b64 "encoding/base64"
	"errors"
	"regexp"
	"time"

	"github.com/hm-edu/portal-common/model"
	"github.com/labstack/echo/v4"
	"github.com/smallstep/certificates/acme"
)

// EabRequest holds the (optional) comment of the token to create.
type EabRequest struct {
	Comment string `json:"comment"`
}

// Bind binds an incoming echo request to the internal model and perfoms a validation
func (r *EabRequest) Bind(c echo.Context, _ *model.Validator) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	match := regexp.MustCompile(`^[a-zA-Z0-9-_.: üäöÄÖÜß]*$`).MatchString(r.Comment)
	if !match {
		return errors.New("comment must match regex ^[a-zA-Z0-9-_.:]*$")
	}
	return nil
}

// EAB is an external account binding.
type EAB struct {
	ID       string     `json:"id"`
	KeyBytes string     `json:"key_bytes"`
	Bound    *time.Time `json:"bound_at"`
	Comment  string     `json:"comment"`
}

// NewEAB creates a new EAB.
func NewEAB(key *acme.ExternalAccountKey) *EAB {
	var bound *time.Time
	if !key.BoundAt.IsZero() {
		bound = &key.BoundAt
	}
	return &EAB{
		ID:       key.ID,
		KeyBytes: b64.RawURLEncoding.EncodeToString(key.HmacKey),
		Bound:    bound,
	}
}
