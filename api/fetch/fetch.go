package fetch

import (
	"github.com/fruitspace/HyprrSpace/api/ent"
	"github.com/fruitspace/HyprrSpace/models/structs"
	"github.com/gofiber/fiber/v2"
	"math"
)

type FetchAPI struct {
	*ent.API
}

func (api *FetchAPI) Register(router fiber.Router) error {
	router.Get("/stats", api.Stats)
	router.Get("/bot_discord", api.DiscordUsers)
	router.Get("/bot_discord_info", api.DiscordUserInfo)

	// GDPS
	router.Get("/gd/tariffs", api.GDTariffs)
	router.Get("/gd/top", api.GDTopServers)
	router.Get("/gd/info/:srvid", api.GDServerInfo)
	router.Get("/partner/obeygdps/:user", api.GDServersByDiscordObey)

	// Minecraft
	router.Get("/mc/cores", api.MinecraftCores)
	router.Get("/mc/regions", api.GetRegions)
	router.Get("/mc/pricing/:region", api.GetPricing)
	return nil
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
