package api

import (
	"errors"
	"github.com/fruitspace/FiberAPI/api/ent"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/gofiber/fiber/v2"
)

type ServersAPI struct {
	*ent.API
}

func (api *ServersAPI) Register(router fiber.Router) error {
	router.Get("/", api.ListServers)
	router.Post("/gd", api.CreateGD)
	router.Post("/mc", api.CreateMC)

	return nil
}

// ServersList returns list of user servers
// @Tags Server Management
// @Summary Returns list of user servers
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Success 200 {object} structs.APIServerListResponse
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /servers [get]
func (api *ServersAPI) ListServers(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	return c.JSON(structs.APIServerListResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		GD:              api.ServerGDProvider.GetUserServers(acc.Data().UID),
		MC:              api.ServerMCProvider.GetUserServers(acc.Data().UID),
		CS:              nil,
	})
}

// ServersCreateGD returns list of user servers
// @Tags Server Management
// @Summary Creates or updates existing GDPS
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Param data body structs.APIServerGDCreateRequest true "Name is used for new server creation, while srvid is used for upgrading existing servers"
// @Success 200 {object} structs.APIBasicSuccess
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /servers/gd [post]
func (api *ServersAPI) CreateGD(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	var data structs.APIServerGDCreateRequest
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}

	api.SuperLock.Lock("gdps_create")
	defer api.SuperLock.Unlock("gdps_create")

	var err error
	var srvid string
	srv := api.ServerGDProvider.New()
	if len(data.SrvId) == 4 {
		err = srv.UpgradeServer(acc.Data().UID, data.SrvId, data.Tariff, data.Duration, data.Promocode)
		srvid = data.SrvId
	} else {
		srvid, err = srv.CreateServer(acc.Data().UID, data.Name, data.Tariff, data.Duration, data.Promocode)
	}
	if err != nil {
		return c.JSON(structs.NewDecoupleAPIError(err))
	}
	return c.JSON(structs.NewAPIBasicResponse(srvid))
}

func (api *ServersAPI) CreateMC(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	var data structs.APIServerMCCreateRequest
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}

	api.SuperLock.Lock("mc_create")
	defer api.SuperLock.Unlock("mc_create")

	var err error
	var srvid string
	srv := api.ServerMCProvider.New()
	srvid, err = srv.CreateServer(
		acc.Data(), data.Name, data.Tariff, data.Core, data.Version, data.AddStorage, data.DedicatedPort, data.Promocode,
	)
	if err != nil {
		return c.JSON(structs.NewDecoupleAPIError(errors.New(srvid + " |" + err.Error())))
	}
	return c.JSON(structs.NewAPIBasicResponse(srvid))
}
