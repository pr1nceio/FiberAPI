package api

import (
	"github.com/fruitspace/FiberAPI/api/ent"
	"github.com/fruitspace/FiberAPI/models/db"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/gofiber/fiber/v2"
	"log"
	"strings"
)

type LegacyAPI struct {
	*ent.API
}

func (api *LegacyAPI) Register(router fiber.Router) error {
	router.Get("/repatch/getserverinfo", api.RepatchGetServerInfo)
	router.Post("/repatch/report", api.RepatchUploadTelemetry)
	return nil
}

func (api *LegacyAPI) RepatchGetServerInfo(c *fiber.Ctx) error {
	srvid := c.Query("id")
	server := api.ServerGDProvider.New().GetReducedServer(srvid)
	if server.SrvName == "" {
		return c.Status(404).JSON(structs.NewAPIError("Server not found"))
	}
	textures := "gdps_textures.zip"
	if server.IsCustomTextures {
		textures = server.SrvID + ".zip"
	}
	data := struct {
		Name        string `json:"name"`
		SrvId       string `json:"srvid"`
		Players     int    `json:"players"`
		Levels      int    `json:"levels"`
		Icon        string `json:"icon"`
		TexturePack string `json:"texturepack"`
		Region      string `json:"region"`
	}{
		server.SrvName,
		server.SrvID,
		server.UserCount,
		server.LevelCount,
		server.Icon,
		textures,
		"ru",
	}
	return c.JSON(data)
}

func (api *LegacyAPI) RepatchUploadTelemetry(c *fiber.Ctx) error {
	type Fingerprint struct {
		CPU   string   `json:"cpu"`
		Cores int      `json:"cores"`
		OS    string   `json:"os"`
		RAM   int      `json:"ram"`
		GPUs  []string `json:"gpu"`
		GUID  string   `json:"guid"`
		AV    string   `json:"av"`
	}
	data := Fingerprint{}
	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(structs.NewAPIError("Invalid request"))
	}
	gpus := strings.Trim(strings.Join(data.GPUs, "\n"), "\n")
	err := api.ServerGDProvider.ExposeGorm().Model(&db.MachineTelemetry{}).Save(db.MachineTelemetry{
		GUID:  data.GUID,
		Os:    data.OS,
		CPU:   data.CPU,
		Cores: data.Cores,
		RAM:   data.RAM,
		Gpus:  gpus,
		Av:    data.AV,
	}).Error
	if err != nil {
		log.Println(err)
	}
	return c.SendString("OK")
}
