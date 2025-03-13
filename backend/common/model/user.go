package model

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// User contains the basic user data provided using OIDC/OAuth2 Access Tokens
type User struct {
	FirstName             string `validate:"required"`
	LastName              string `validate:"required"`
	MiddleName            string
	Email                 string `validate:"required"`
	AdditionalSmimeEmails []string
	CommonName            string `validate:"required"`
	Student               bool
}

// Bind binds an incoming echo request to the the User and perfoms a validation
func (r *User) Bind(c echo.Context, v *Validator) error {
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)

	r.FirstName = claims["given_name"].(string)
	r.LastName = claims["family_name"].(string)
	r.CommonName = claims["name"].(string)
	r.Email = claims["email"].(string)
	if claim, ok := claims["additional_smime_emails"].([]interface{}); ok {
		for _, v := range claim {
			if email, ok := v.(string); ok {
				r.AdditionalSmimeEmails = append(r.AdditionalSmimeEmails, email)
			}
		}
	}
	if claim, ok := claims["eduPersonScopedAffiliation"]; ok {
		if affiliations, ok := claim.([]string); ok {
			for _, affiliation := range affiliations {
				if strings.Contains(affiliation, "student@") {
					r.Student = true
				}
			}
		} else if affiliation, ok := claim.(string); ok {
			if strings.Contains(affiliation, "student@") {
				r.Student = true
			}
		}
	} else {
		r.Student = false
	}
	err := v.Validate(r)
	return err
}
