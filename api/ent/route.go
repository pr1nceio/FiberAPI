package ent

import (
	"github.com/gofiber/fiber/v2"
	"log"
)

type ApiRoute interface {
	Register(router fiber.Router) error
}

func Register(api ApiRoute, router fiber.Router) {
	if err := api.Register(router); err != nil {
		log.Panicln(err)
	}
}
