package model

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

type User struct {
	Firstname  string `validate:"required"`
	Lastname   string `validate:"required"`
	Email      string `validate:"required"`
	CommonName string `validate:"required"`
}

// Bind binds an incoming echo request to the the User and perfoms a validation
func (r *User) Bind(c echo.Context, v *Validator) error {
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	r = &User{
		Firstname:  claims["given_name"].(string),
		Lastname:   claims["family_name"].(string),
		CommonName: claims["name"].(string),
		Email:      claims["email"].(string),
	}
	err := v.Validate(r)
	return err
}
