package api

import (
	"fmt"
	"github.com/fruitspace/FiberAPI/api/ent"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"log"
	"runtime/debug"
	"strings"
)

func StartServer(api ent.API) error {
	app := fiber.New(fiber.Config{
		BodyLimit:     5 * 1024 * 1024, // 5MB Body Limit?
		CaseSensitive: true,
		ServerHeader:  "Fiber",
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     "https://fruitspace.ru, https://fruitspace.one, https://gofruit.space",
	}))
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
	app.Group("/antiswagger").Use(basicauth.New(basicauth.Config{Users: map[string]string{
		"dev": "fruitspace_swag",
	}})).Get("/*", swagger.HandlerDefault) //Swag

	app.All("/", shield)

	//region Auth
	auth := app.Group("/auth")
	ent.Register(&AuthAPI{&api}, auth)
	//endregion

	//region User
	user := app.Group("/user")
	ent.Register(&UserAPI{&api}, user)
	// endregion

	//region Payments
	payments := app.Group("/payments")
	ent.Register(&PaymentsAPI{&api}, payments)
	//endregion

	//region Fetch
	fetch := app.Group("/fetch")
	ent.Register(&FetchAPI{&api}, fetch)
	//endregion

	//region Admin
	admin := app.Group("/admin")
	ent.Register(&AdminAPI{&api}, admin)
	//endregion

	// region Servers
	servers := app.Group("/servers")
	ent.Register(&ServersAPI{&api}, servers)
	// endregion

	//region GDPS
	gdps := servers.Group("/gd/:srvid")
	ent.Register(&ManageGDPSAPI{&api}, gdps)

	gdps_user := gdps.Group("/u")
	ent.Register(&AuxiliaryGDPSAPI{&api}, gdps_user)
	//endregion

	// region Minecraft
	//mc := servers.Group("/mc/:srvid")
	//mc.Get("/", nil)       // get server
	//mc.Delete("/", nil) //delete
	// endregion

	// region Internal
	internal := app.Group("/internal")
	ent.Register(&InternalGDPSApi{&api}, internal)
	// endregion

	//region Legacy
	legacy := app.Group("/v1")
	ent.Register(&LegacyAPI{&api}, legacy)
	//endregion

	// region Particle
	particle := app.Group("/particle")
	ent.Register(&ParticleAPI{&api}, particle)
	// endregion

	// region Repatch
	repatch := app.Group("/repatch")
	ent.Register(&RepatchAPI{&api}, repatch)
	// endregion

	routes := collectRoutes(app)
	app.Get("/highlyadvertisedendpoint", func(c *fiber.Ctx) error {
		return c.SendString(strings.Join(routes, "\n"))
	})

	return app.Listen(api.Host)
}

func collectRoutes(app *fiber.App) []string {
	var routes []string
	for _, route := range app.GetRoutes(true) {
		routes = append(routes, fmt.Sprintf("%s %s", route.Method, route.Path))
	}
	return routes
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

func shield(c *fiber.Ctx) error {
	return c.Status(200).JSON(structs.NewAPIBasicResponse("FiberAPI is alive"))
}
