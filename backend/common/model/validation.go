package model

import "github.com/go-playground/validator/v10"

// Validator is a wrapper around a go validator.
type Validator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator.
func NewValidator() *Validator {
	return &Validator{
		validator: validator.New(),
	}
}

// Validate performs a validation using the attached validator.
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}
