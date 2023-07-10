package api

import (
	"encoding/json"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/gofiber/fiber/v2"
	"log"
)

func (api *API) APIGDPSSendWebhook(c *fiber.Ctx) error {
	srvid := c.Params("srvid")
	xtype := c.Query("type")
	var data map[string]string
	body := c.Request().Body()
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Println(err)
		log.Println(string(body))
		return c.SendString("A")
	}
	srv := api.ServerGDProvider.New()
	if len(srvid) != 4 || !srv.Exists(srvid) {
		return c.Status(500).JSON(structs.NewAPIError("No server found"))
	}
	srv.GetServerBySrvID(srvid)
	srv.SendWebhook(xtype, data)
	return c.SendString("OK")
}
