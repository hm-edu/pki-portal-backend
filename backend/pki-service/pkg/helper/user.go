package helper

import (
	"crypto"

	"github.com/go-acme/lego/v4/registration"
)

// User represents an ACME user.
type User struct {
	Email        string
	Registration *registration.Resource
	Key          crypto.PrivateKey
}

// GetEmail returns the email of the user.
func (u *User) GetEmail() string {
	return u.Email
}

// GetRegistration returns the registration resource of the user.
func (u User) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey returns the private key of the user.
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.Key
}
