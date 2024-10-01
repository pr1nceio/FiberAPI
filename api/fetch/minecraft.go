package fetch

import (
	"github.com/fruitspace/HyprrSpace/models/db"
	"github.com/fruitspace/HyprrSpace/models/structs"
	"github.com/gofiber/fiber/v2"
)

func (api *FetchAPI) MinecraftCores(c *fiber.Ctx) error {
	return c.JSON(structs.MinecraftCoresResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Cores:           structs.MCCoresEggs,
	})
}

func (api *FetchAPI) GetRegions(c *fiber.Ctx) error {
	regions := api.ServerMCProvider.ListRegions()
	return c.JSON(struct {
		structs.APIBasicSuccess
		Regions []db.RegionPublic `json:"regions"`
	}{
		structs.NewAPIBasicResponse("Success"),
		regions,
	})
}

func (api *FetchAPI) GetPricing(c *fiber.Ctx) error {
	region, err := c.ParamsInt("region")
	if err != nil || region == 0 {
		return c.Status(400).JSON(structs.NewAPIError("Invalid region"))
	}
	tariffs := api.ServerMCProvider.GetPricing(region)
	return c.JSON(struct {
		structs.APIBasicSuccess
		Tariffs []db.PricingPublic `json:"tariffs"`
	}{
		structs.NewAPIBasicResponse("Success"),
		tariffs,
	})
}
