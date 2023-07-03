package api

import (
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/providers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"strings"
)

type API struct {
	AccountProvider      *providers.AccountProvider
	NotificationProvider *providers.NotificationProvider
	PaymentProvider      *providers.PaymentProvider
	PromocodeProvider    *providers.PromocodeProvider
	ShopProvider         *providers.ShopProvider
	ServerGDProvider     *providers.ServerGDProvider
	Host                 string
}

func StartServer(api API) error {
	app := fiber.New(fiber.Config{
		BodyLimit:     5 * 1024 * 1024,
		CaseSensitive: true,
		ServerHeader:  "Fiber",
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{AllowCredentials: true}))
	app.Get("/antiswagger/*", swagger.HandlerDefault) //Swag

	app.All("/", shield)

	//region Auth
	app.Post("/auth/login", api.AuthLogin)
	app.Post("/auth/register", api.AuthRegister)
	app.All("/auth/confirm_email", api.AuthConfirmEmail)
	app.Post("/auth/recover", api.AuthRecoverPassword)
	app.All("/auth/discord", api.AuthDiscord)
	//endregion

	//region User
	app.Get("/user", api.UserSSO)                  //sso, top_server
	app.Patch("/user", api.UserUpdate)             //change name, password, totp
	app.Post("/user/avatar", api.UserAvatarUpdate) //avatar

	app.Get("/payments", api.PaymentsGet)     // get payments
	app.Post("/payments", api.PaymentsCreate) //create payment
	//endregion

	//region Fetch
	fetch := app.Group("/fetch")
	fetch.Get("/stats", api.FetchStats)
	fetch.Get("/bot_discord", api.FetchDiscordUsers)
	fetch.Get("/bot_discord_info", api.FetchDiscordUserInfo)
	fetch.Get("/gd/tariffs", api.FetchGDTariffs)
	fetch.Get("/gd/top", api.FetchGDTopServers)
	fetch.Get("/gd/info/:srvid", api.FetchGDServerInfo) // get public gdps download card
	//endregion

	servers := app.Group("/servers")
	servers.Get("/", api.ServersList) // list servers

	//region GDPS
	servers.Post("/gd", api.ServersCreateGD) // create
	gdps := servers.Group("/gd/:srvid")
	gdps.Get("/", api.ManageGDPSGet)       // get server
	gdps.Delete("/", api.ManageGDPSDelete) //delete

	gdps.Post("/settings", api.ManageGDPSChangeSettings) //change settings
	gdps.Post("/icon", api.ManageGDPSChangeIcon)         //icon
	gdps.Get("/dbreset", api.ManageGDPSResetDBPassword)  //reset DB Password
	gdps.Post("/chests", api.ManageGDPSUpdateChests)     //update chests
	gdps.Get("/logs", api.ManageGDPSGetLogs)             //get logs by filter
	gdps.Post("/music", api.ManageGDPSGetMusic)          //get songs by filter
	gdps.Put("/music", api.ManageGDPSAddMusic)           //put songs

	//gdps.Get("/buildlab")
	gdps.Post("/buildlab", api.ManageGDPSBuildLabPush)
	//endregion

	return app.Listen(api.Host)
}

func getIP(ctx *fiber.Ctx) string {
	IPAddr := ctx.Get("CF-Connecting-IP")
	if IPAddr == "" {
		IPAddr = ctx.Get("X-Real-IP")
	}
	if IPAddr == "" {
		IPAddr = strings.Split(ctx.IP(), ":")[0]
	}
	return IPAddr
}

func (api *API) performAuth(c *fiber.Ctx, acc *providers.Account) bool {
	token := c.Get("Authorization")
	if token == "" || !acc.GetUserBySession(token) {
		return false
	}
	if !acc.Data().IsActivated || acc.Data().IsBanned {
		return false
	}
	return true
}

func shield(c *fiber.Ctx) error {
	return c.Status(200).JSON(structs.NewAPIBasicResponse("FiberAPI is alive"))
}
