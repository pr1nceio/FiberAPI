package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func RegisterRoutes(app *fiber.App) {
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{AllowCredentials: true}))

}

func StartServer(host string) error {
	app := fiber.New()
	RegisterRoutes(app)
	return app.Listen(host)
}
