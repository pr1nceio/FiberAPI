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
	}{
		"ok",
		api.ServerGDProvider.CountServers(),
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
