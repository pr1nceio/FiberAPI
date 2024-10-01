package api

import (
	"github.com/fruitspace/HyprrSpace/api/ent"
	"github.com/fruitspace/HyprrSpace/models/structs"
	"github.com/gofiber/fiber/v2"
)

type RepatchAPI struct {
	*ent.API
}

func (api *RepatchAPI) Register(router fiber.Router) error {
	router.Get("/gd/:id", api.GDInfo)
	return nil
}

func (api *RepatchAPI) GDInfo(c *fiber.Ctx) error {
	srvid := c.Params("id")
	server := api.ServerGDProvider.New().GetRepatchServer(srvid)
	if server == nil {
		return c.Status(404).JSON(structs.NewAPIError("Server not found"))
	}

	return c.JSON(server)
}
