package api

import (
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/gofiber/fiber/v2"
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
		Status    string `json:"status"`
		GDPSCount int    `json:"gdps_count"`
	}{"ok", api.ServerGDProvider.CountServers()})
}
