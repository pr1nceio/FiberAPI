package api

import (
	"fmt"
	"github.com/fruitspace/FiberAPI/api/ent"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/services"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/gofiber/fiber/v2"
	"log"
	"strconv"
)

type AdminAPI struct {
	*ent.API
}

func (api *AdminAPI) Register(route fiber.Router) error {
	route.Get("/bot_clear", api.CleanUnpaidInstallers)
	return nil
}

func (api *AdminAPI) CleanUnpaidInstallers(c *fiber.Ctx) error {
	if c.Query("key") != "M41dss0nT0p" {
		return c.Status(404).SendString("Cannot GET /admin/bot_clear")
	}
	utils.SendMessageDiscord("Starting forced maintenance...")

	// Check paid GDPS servers expiry

	if c.Query("paid") != "" {
		gdpslist := api.ServerGDProvider.GetUnpaidServers()
		freezeReport := "### Freezing Paid GDPS (" + strconv.Itoa(len(gdpslist)) + ")...\n"
		for _, gdps := range gdpslist {
			srv := api.ServerGDProvider.New()
			if !srv.GetServerBySrvID(gdps) {
				continue
			}
			srv.LoadCoreConfig()
			if srv.CoreConfig.ServerConfig.Locked {
				continue
			}
			srv.FreezeServer()
			freezeReport += fmt.Sprintf("\n❄️ %s is frozen", gdps)
			if len(freezeReport) > 500 {
				utils.SendMessageDiscord(freezeReport)
				freezeReport = ""
			}
		}
		utils.SendMessageDiscord(freezeReport)
	}

	// Purge Empty Free GDPS and Freeze them
	if c.Query("free") != "" {
		gdpslist := api.ServerGDProvider.GetInactiveServers(3, true)
		freezeReport := "### Purging Free Inactive GDPS (" + strconv.Itoa(len(gdpslist)) + ")...\n"
		for _, gdps := range gdpslist {
			log.Println("Processing ", gdps)
			srv := api.ServerGDProvider.New()
			if !srv.GetServerBySrvID(gdps) {
				continue
			}
			srv.LoadCoreConfig()
			interactor := srv.NewInteractor()
			usrs := interactor.CountActiveUsersLastWeek()

			if usrs > 0 {
				continue
			}
			// If there's no active users last week then freeze server and set expire date to current time to ensure removal in future

			if srv.CoreConfig.ServerConfig.Locked && c.Query("force") == "" {
				continue
			}
			srv.FreezeServer()
			err := srv.DeleteInstallers()
			if err != nil {
				freezeReport += fmt.Sprintf("\n❌ Couldn't purge %s, error: `%s`", gdps, err.Error())
			} else {
				freezeReport += fmt.Sprintf("\n❄️ %s is purged", gdps)
			}
			if len(freezeReport) > 500 {
				utils.SendMessageDiscord(freezeReport)
				freezeReport = ""
			}
		}
		utils.SendMessageDiscord(freezeReport)
	}

	// Fix missing installers
	if c.Query("installers") != "" {
		gdpslist := api.ServerGDProvider.GetMissingInstallersServers()
		freezeReport := "### Restoring GDPS with missing installers (" + strconv.Itoa(len(gdpslist)) + ")...\n"
		for _, gdps := range gdpslist {
			srv := api.ServerGDProvider.New()
			if !srv.GetServerBySrvID(gdps) {
				continue
			}
			srv.LoadCoreConfig()

			if srv.CoreConfig.ServerConfig.Locked {
				continue
			}
			err := srv.ExecuteBuildLab(structs.BuildLabSettings{
				SrvName:  srv.Srv.SrvName,
				Version:  "2.1",
				Windows:  true,
				Android:  true,
				IOS:      false,
				MacOS:    false,
				Icon:     "gd_default.png",
				Textures: "default",
			})
			if err != nil {
				freezeReport += fmt.Sprintf("\n❌ Couldn't recover %s, error: `%s`", gdps, err.Error())
			} else {
				freezeReport += fmt.Sprintf("\n❄⚙️ %s is recovered W+A", gdps)
			}
			if len(freezeReport) > 500 {
				utils.SendMessageDiscord(freezeReport)
				freezeReport = ""
			}
		}
		utils.SendMessageDiscord(freezeReport)
	}

	//Clear music
	mus := services.InitMusic(api.ServerGDProvider.ExposeRedis(), "admin")
	mcnt := mus.CleanEmptyNewgrounds()
	utils.SendMessageDiscord(fmt.Sprintf("Cleaned %d invalid NG songs. \n### Maintenance Complete", mcnt))

	return c.SendString("OK")
}
