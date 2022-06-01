package models

import (
	"time"

	"github.com/smallstep/certificates/acme"
)

// EAB is an external account binding.
type EAB struct {
	ID       string     `json:"id"`
	KeyBytes []byte     `json:"key_bytes"`
	Bound    *time.Time `json:"bound_at"`
}

// NewEAB creates a new EAB.
func NewEAB(key *acme.ExternalAccountKey) *EAB {
	var bound *time.Time
	if !key.BoundAt.IsZero() {
		bound = &key.BoundAt
	}
	return &EAB{
		ID:       key.ID,
		KeyBytes: key.HmacKey,
		Bound:    bound,
	}
}
