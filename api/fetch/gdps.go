package fetch

import (
	"fmt"
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"time"
)

// FetchGDTariffs returns a list of available GDPS Tariffs [/fetch/gd/tariffs]
func (api *FetchAPI) GDTariffs(c *fiber.Ctx) error {
	return c.JSON(struct {
		Status  string `json:"status"`
		Tariffs map[string]structs.GDTariff
	}{"ok", fiberapi.ProductGDTariffs})
}

// GDServerInfo returns an eligible download page, etc. data
// @Tags Public stats fetching
// @Summary Returns
// @Accept json
// @Produce json
// @Param srvid path string true "GDPS Server ID"
// @Response 200 {object} db.ServerGDReduced
// @Router /fetch/gd/info/{srvid} [get]
func (api *FetchAPI) GDServerInfo(c *fiber.Ctx) error {
	gdpsID := c.Params("srvid")
	server := api.ServerGDProvider.New().GetReducedServer(gdpsID)
	return c.JSON(server)
}

func (api *FetchAPI) GDServersByDiscordObey(c *fiber.Ctx) error {
	discordId := c.Params("user")
	partnerKey := c.Query("partner_key")
	if partnerKey != "En1gM4t0s" {
		return c.Status(fiber.StatusNotFound).SendString(fmt.Sprintf("Cannot %s %s", c.Method(), c.Path()))
	}
	user := api.AccountProvider.New()
	if !user.GetUserByDiscord(discordId) {
		return c.JSON(structs.NewAPIError("No User"))
	}
	servers := api.ServerGDProvider.GetUserServers(user.Data().UID)
	t := time.Now()
	for _, s := range servers {
		//cleanup
		s.Plan = 0
		s.ExpireDate = t
		s.OwnerID = 0
	}
	return c.JSON(structs.ObeyGDPSResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Username:        user.Data().Uname,
		ProfilePic:      user.Data().ProfilePic,
		Servers:         servers,
	})
}

func (api *FetchAPI) GDTopServers(c *fiber.Ctx) error {
	offsetp := c.Query("offset")
	offset, _ := strconv.Atoi(offsetp)
	servers := api.ServerGDProvider.GetTopServers(offset)
	return c.JSON(structs.APITopGDServers{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Servers:         servers,
	})
}
