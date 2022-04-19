package model

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

// User contains the basic user data provided using OIDC/OAuth2 Access Tokens
type User struct {
	FirstName  string `validate:"required"`
	LastName   string `validate:"required"`
	MiddleName string
	Email      string `validate:"required"`
	CommonName string `validate:"required"`
}

// Bind binds an incoming echo request to the the User and perfoms a validation
func (r *User) Bind(c echo.Context, v *Validator) error {
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	r = &User{
		FirstName:  claims["given_name"].(string),
		LastName:   claims["family_name"].(string),
		CommonName: claims["name"].(string),
		Email:      claims["email"].(string),
	}
	err := v.Validate(r)
	return err
}
