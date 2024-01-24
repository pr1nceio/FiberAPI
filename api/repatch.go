package api

import (
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/gofiber/fiber/v2"
)

func (api *API) RepatchGDInfo(c *fiber.Ctx) error {
	srvid := c.Params("id")
	server := api.ServerGDProvider.New().GetRepatchServer(srvid)
	if server == nil {
		return c.Status(404).JSON(structs.NewAPIError("Server not found"))
	}

	return c.JSON(server)
}
