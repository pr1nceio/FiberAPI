package api

import (
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/gofiber/fiber/v2"
)

func (api *API) ParticleSearch(c *fiber.Ctx) error {
	var req structs.APIParticleSearchRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(structs.NewAPIError("invalid json", err.Error()))
	}
	particles, count, err := api.ParticleProvider.SearchParticles(req.Query, req.Arch, req.IsOfficial, req.Sort, req.Page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(structs.NewDecoupleAPIError(err))
	}
	return c.Status(fiber.StatusOK).JSON(structs.APIParticleSearchResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("success"),
		Particles:       particles,
		Count:           count,
	})
}

func (api *API) ParticleGet(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	if !api.performAuth(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	author := c.Params("author")
	name := c.Params("name")

	particle, err := api.ParticleProvider.GetParticle(name, author, acc.Data().UID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(structs.NewDecoupleAPIError(err))
	}
	particle.APIBasicSuccess = structs.NewAPIBasicResponse("success")
	return c.Status(fiber.StatusOK).JSON(particle)
}

func (api *API) ParticleGetUser(c *fiber.Ctx) error {
	q := c.QueryBool("reg")
	acc := api.AccountProvider.New()
	if !api.performAuth(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	pu := api.ParticleProvider.NewUser()
	err := pu.GetByUID(acc.Data().UID)
	if err != nil {
		if q {
			err = pu.RegisterFromUser(acc.Data())
			if err != nil {
				return c.Status(500).JSON(structs.NewAPIError(err.Error()))
			}
			return c.Status(200).JSON(structs.APIParticleUserResponse{
				APIBasicSuccess: structs.NewAPIBasicResponse("success"),
				ParticleUser:    *pu.Data,
			})
		} else {
			return c.Status(500).JSON(structs.NewAPIError(err.Error()))
		}
	}
	used, err := pu.CalculateUsedSize()
	if err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	if used == nil {
		used = new(uint)
	}
	return c.Status(200).JSON(structs.APIParticleUserResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("success"),
		ParticleUser:    *pu.Data,
		UsedSize:        *used,
	})
}
