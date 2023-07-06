package api

import (
	"fmt"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/gofiber/fiber/v2"
)

func (api *API) AdminCleanUnpaidInstallers(c *fiber.Ctx) error {
	if c.Query("key") != "M41dss0nT0p" {
		return c.Status(404).SendString("Cannot GET /admin/bot_clear")
	}
	gdpslist := api.ServerGDProvider.GetUnpaidServers()
	freezeReport := ""
	for _, gdps := range gdpslist {
		srv := api.ServerGDProvider.New()
		if !srv.GetServerBySrvID(gdps) {
			continue
		}
		srv.FreezeServer()
		srv.DeleteInstallers()
		freezeReport += fmt.Sprintf("\n❄️ %s is frozen (and wiped)", gdps)
		if len(freezeReport) > 500 {
			utils.SendMessageDiscord(freezeReport)
			freezeReport = ""
		}
	}
	utils.SendMessageDiscord(freezeReport)

	return c.SendString("OK")
}
