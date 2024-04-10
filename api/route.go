package api

import "github.com/gofiber/fiber/v2"

type Route interface {
	Register(router fiber.Router) error
}
