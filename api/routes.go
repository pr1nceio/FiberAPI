package api

import (
	"github.com/fruitspace/FiberAPI/providers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
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
	app.Use(cors.New(cors.Config{AllowCredentials: true}))
	app.Get("/antiswagger/*", swagger.HandlerDefault) //Swag

	//region Auth
	app.Post("/auth/login", api.AuthLogin)
	app.Post("/auth/register", api.AuthRegister)
	app.All("/auth/confirm_email", api.AuthConfirmEmail)
	app.Post("/auth/recover", api.AuthRecoverPassword)
	//endregion

	////region User
	//app.Get("/user")         //sso, top_server
	//app.Patch("/user")       //change name, password, totp
	//app.Post("/user/avatar") //avatar
	//
	//app.Get("/payments")  // get payments
	//app.Post("/payments") //create payment
	////endregion

	//region Fetch
	fetch := app.Group("/fetch")
	fetch.Get("/stats")
	fetch.Get("/gd/tariffs", api.FetchGDTariffs)
	fetch.Get("/gd/info/:srvid") // get public gdps download card
	//endregion

	//servers := app.Group("/servers")
	//servers.Get("/") // list servers
	//
	////region GDPS
	//servers.Post("/gd") // create
	//gdps := servers.Group("/gd/:srvid")
	//gdps.Get("/")    // get server
	//gdps.Delete("/") //delete
	//
	//gdps.Post("/settings") //change settings
	//gdps.Post("/icon")     //icon
	//gdps.Get("/dbreset")   //reset DB Password
	//gdps.Get("/chests")
	//gdps.Post("/chests")
	//gdps.Get("/logs")
	//gdps.Get("/music")
	//gdps.Post("/music")
	//
	//gdps.Get("/buildlab")
	//gdps.Post("/buildlab")
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
