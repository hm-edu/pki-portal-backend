package models

import "github.com/smallstep/certificates/acme"

// EAB is an external account binding.
type EAB struct {
	ID       string `json:"id"`
	KeyBytes []byte `json:"keyBytes"`
}

// NewEAB creates a new EAB.
func NewEAB(key *acme.ExternalAccountKey) *EAB {
	return &EAB{
		ID:       key.ID,
		KeyBytes: key.HmacKey,
	}
}
