package models

import "github.com/gofiber/fiber/v2"

type ErrorHandler = func(*fiber.Ctx, Error) error

type Error struct {
	Status  int `json:"-"`
	Message string
	Details string
}
