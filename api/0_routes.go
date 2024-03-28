package api

import (
	"fmt"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/providers"
	"github.com/fruitspace/FiberAPI/providers/ServerGD"
	"github.com/fruitspace/FiberAPI/providers/ServerMC"
	"github.com/fruitspace/FiberAPI/providers/particle"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"log"
	"runtime/debug"
	"strings"
)

type API struct {
	AccountProvider      *providers.AccountProvider
	NotificationProvider *providers.NotificationProvider
	PaymentProvider      *providers.PaymentProvider
	PromocodeProvider    *providers.PromocodeProvider
	ShopProvider         *providers.ShopProvider
	ServerGDProvider     *ServerGD.ServerGDProvider
	ServerMCProvider     *ServerMC.ServerMCProvider
	ParticleProvider     *particle.ParticleProvider

	SuperLock *utils.SuperLock
	Host      string
}

func StartServer(api API) error {
	app := fiber.New(fiber.Config{
		BodyLimit:     5 * 1024 * 1024, // 5MB Body Limit?
		CaseSensitive: true,
		ServerHeader:  "Fiber",
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{AllowCredentials: true}))
	app.Use(recover.New(recover.Config{
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			go sentry.CaptureException(e.(error))
			log.Println(string(debug.Stack()))
			utils.SendMessageDiscord(fmt.Sprintf("[%s] Got panic at FiberAPI, check logs\n<@886130124225937409>",
				utils.GetEnv("NOMAD_SHORT_ALLOC_ID", "default")))
		},
		EnableStackTrace: true,
	}))
	app.Use(pprof.New(pprof.Config{
		Prefix: "/highlyadvertisedendpointplsdontdoanythingbad",
	}))
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
	app.Get("/user/joinguild", api.UserJoinGuild)  //avatar
	app.Get("/user/sessions", api.UserListSessions)
	app.Delete("/user/sessions", api.UserDeleteSessions)

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

	fetch.Get("/mc/cores", api.FetchMinecraftCores)
	//endregion

	//region Admin
	admin := app.Group("/admin")
	admin.Get("/bot_clear", api.AdminCleanUnpaidInstallers)
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
	gdps.Post("/logs", api.ManageGDPSGetLogs)            //get logs by filter
	gdps.Post("/music", api.ManageGDPSGetMusic)          //get songs by filter
	gdps.Put("/music", api.ManageGDPSAddMusic)           //put songs
	gdps.Get("/roles", api.ManageGDPSGetRoles)           //get roles
	gdps.Post("/roles", api.ManageGDPSSetRole)           //create or update role
	gdps.Get("/levelpacks", api.ManageGDPSGetLevelPacks) //get levelpacks

	gdps.Get("/get/users", api.ManageGDPSQueryUsers)   //get users
	gdps.Get("/get/levels", api.ManageGDPSQueryLevels) //get levels

	gdps.Put("/modules/discord", api.ManageGDPSDiscordModule)

	gdps.Get("/upgrade22", api.ManageGDPSPressStart22Upgrade)
	gdps.Post("/buildlab", api.ManageGDPSBuildLabPush)
	gdps.Get("/buildlab/status", api.ManageGDPSGetBuildStatus)

	gdps_user := gdps.Group("/u")
	gdps_user.Post("/login", api.AuxiliaryGDPSLogin)
	gdps_user.Get("/", api.AuxiliaryGDPSAuth)
	gdps_user.Put("/", api.AuxiliaryGDPSChangeCreds)
	gdps_user.Post("/music", api.AuxiliaryGDPSGetMusic)
	gdps_user.Put("/music", api.AuxiliaryGDPSAddMusic)
	gdps_user.Post("/recover", api.AuxiliaryGDPSForgotPassword)
	//endregion

	// region Minecraft
	servers.Post("/mc", api.ServersCreateMC) // Create
	//mc := servers.Group("/mc/:srvid")
	//mc.Get("/", nil)       // get server
	//mc.Delete("/", nil) //delete
	// endregiom

	internal := app.Group("/internal")
	internal.Post("/gd/:srvid/webhook", api.APIGDPSSendWebhook)

	//region Legacy
	app.Get("/v1/repatch/getserverinfo", api.LegacyRepatchGetServerInfo)
	app.Post("/v1/repatch/report", api.LegacyRepatchUploadTelemetry)
	app.Get("/v1/auth/confirm_email", api.AuthConfirmEmail)
	//endregion

	// region Particle
	app.Post("/particle/search", api.ParticleSearch)
	app.Get("/particle/user", api.ParticleGetUser)
	app.Get("/particle/v/:author/:name", api.ParticleGet)
	// endregion

	// region Repatch
	app.Get("/repatch/gd/:id", api.RepatchGDInfo)
	// endregion

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

func getUserAgent(ctx *fiber.Ctx) string {
	ua := ctx.Get("User-Agent", "Unknown")
	return utils.ParseUserAgent(ua)
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
