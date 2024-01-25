package api

import (
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/gofiber/fiber/v2"
	"math"
	"strconv"
)

// FetchGDTariffs returns a list of available GDPS Tariffs [/fetch/gd/tariffs]
func (api *API) FetchGDTariffs(c *fiber.Ctx) error {
	return c.JSON(struct {
		Status  string `json:"status"`
		Tariffs map[string]structs.GDTariff
	}{"ok", fiberapi.ProductGDTariffs})
}

// FetchStats returns all server type count stats
func (api *API) FetchStats(c *fiber.Ctx) error {
	return c.JSON(struct {
		Status     string `json:"status"`
		Clients    int    `json:"clients"`
		GDPSCount  int    `json:"gdps_count"`
		GDPSlevels int    `json:"gdps_levels"`
	}{
		"ok",
		api.AccountProvider.GetUserCount(),
		api.ServerGDProvider.CountServers(),
		api.ServerGDProvider.CountLevels(),
	})
}

// FetchGDServerInfo returns an eligible download page, etc. data
// @Tags Public stats fetching
// @Summary Returns
// @Accept json
// @Produce json
// @Param srvid path string true "GDPS Server ID"
// @Response 200 {object} db.ServerGdReduced
// @Router /fetch/gd/info/{srvid} [get]
func (api *API) FetchGDServerInfo(c *fiber.Ctx) error {
	gdpsID := c.Params("srvid")
	server := api.ServerGDProvider.New().GetReducedServer(gdpsID)
	return c.JSON(server)
}

func (api *API) FetchGDTopServers(c *fiber.Ctx) error {
	offsetp := c.Query("offset")
	offset, _ := strconv.Atoi(offsetp)
	servers := api.ServerGDProvider.GetTopServers(offset)
	return c.JSON(structs.APITopGDServers{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Servers:         servers,
	})
}

func (api *API) FetchDiscordUsers(c *fiber.Ctx) error {
	accs := api.AccountProvider.GetDiscordIntegrations(false)
	accsr := api.AccountProvider.GetDiscordIntegrations(true)
	return c.JSON(map[string][]string{
		"clients": accs,
		"vip":     accsr,
	})
}

func (api *API) FetchDiscordUserInfo(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	if !acc.GetUserByDiscord(c.Query("discord_id", "-1")) {
		return c.JSON(structs.NewAPIError("No User"))
	}
	active := "🟢"
	if !acc.Data().IsActivated {
		active = "🟡"
	}
	if acc.Data().IsBanned {
		active = "🔴"
	}
	servers := api.ServerGDProvider.GetUserServers(acc.Data().UID)
	//SELECT COUNT(*) as pos FROM servers WHERE userCount >= CURRENT
	return c.JSON(structs.APIFetchDiscordUser{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		UID:             acc.Data().UID,
		Uname:           acc.Data().Uname,
		Avatar:          acc.Data().ProfilePic,
		Active:          active,
		Balance:         int(math.Floor(acc.Data().Balance)),
		Servers:         servers,
	})
}

func (api *API) FetchMinecraftCores(c *fiber.Ctx) error {
	return c.JSON(structs.MinecraftCoresResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Cores:           structs.MCCoresEggs,
	})
}
