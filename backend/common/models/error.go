package models

import "github.com/gofiber/fiber/v2"

// ErrorHandler is a type defintion for errors caused by middlewares.
type ErrorHandler = func(*fiber.Ctx, Error) error

// Error defines the (json) structure of a API error.
type Error struct {
	Status  int `json:"-"`
	Message string
	Details string
}
