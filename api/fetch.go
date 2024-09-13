package api

import (
	"fmt"
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/api/ent"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/gofiber/fiber/v2"
	"math"
	"strconv"
	"time"
)

type FetchAPI struct {
	*ent.API
}

func (api *FetchAPI) Register(router fiber.Router) error {
	router.Get("/stats", api.Stats)
	router.Get("/bot_discord", api.DiscordUsers)
	router.Get("/bot_discord_info", api.DiscordUserInfo)
	router.Get("/gd/tariffs", api.GDTariffs)
	router.Get("/gd/top", api.GDTopServers)
	router.Get("/gd/info/:srvid", api.GDServerInfo)                   // get public gdps download card
	router.Get("/partner/obeygdps/:user", api.GDServersByDiscordObey) // get public gdps download card
	router.Get("/mc/cores", api.MinecraftCores)
	return nil
}

// FetchGDTariffs returns a list of available GDPS Tariffs [/fetch/gd/tariffs]
func (api *FetchAPI) GDTariffs(c *fiber.Ctx) error {
	return c.JSON(struct {
		Status  string `json:"status"`
		Tariffs map[string]structs.GDTariff
	}{"ok", fiberapi.ProductGDTariffs})
}

// FetchStats returns all server type count stats
func (api *FetchAPI) Stats(c *fiber.Ctx) error {
	return c.JSON(struct {
		Status     string `json:"status"`
		Clients    int    `json:"clients"`
		GDPSCount  int    `json:"gdps_count"`
		GDPSlevels int    `json:"gdps_levels"`
		MCCount    int    `json:"mc_count"`
	}{
		"ok",
		api.AccountProvider.GetUserCount(),
		api.ServerGDProvider.CountServers(),
		api.ServerGDProvider.CountLevels(),
		api.ServerMCProvider.CountServers(),
	})
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

func (api *FetchAPI) DiscordUsers(c *fiber.Ctx) error {
	accs := api.AccountProvider.GetDiscordIntegrations(false)
	accsr := api.AccountProvider.GetDiscordIntegrations(true)
	return c.JSON(map[string][]string{
		"clients": accs,
		"vip":     accsr,
	})
}

func (api *FetchAPI) DiscordUserInfo(c *fiber.Ctx) error {
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

func (api *FetchAPI) MinecraftCores(c *fiber.Ctx) error {
	return c.JSON(structs.MinecraftCoresResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Cores:           structs.MCCoresEggs,
	})
}
