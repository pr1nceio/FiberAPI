package api

import (
	"encoding/json"
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/api/ent"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/providers"
	"github.com/fruitspace/FiberAPI/providers/ServerGD"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/slices"
	"io"
	"net/http"
	"runtime/debug"
)

type ManageGDPSAPI struct {
	*ent.API
}

func (api *ManageGDPSAPI) Register(router fiber.Router) error {
	router.Get("/", api.Get)       // get server
	router.Delete("/", api.Delete) //delete

	router.Post("/settings", api.ChangeSettings) //change settings
	router.Post("/icon", api.ChangeIcon)         //icon
	router.Get("/dbreset", api.ResetDBPassword)  //reset DB Password
	router.Post("/chests", api.UpdateChests)     //update chests
	router.Post("/logs", api.GetLogs)            //get logs by filter
	router.Post("/music", api.GetMusic)          //get songs by filter
	router.Put("/music", api.AddMusic)           //put songs
	router.Get("/roles", api.GetRoles)           //get roles
	router.Post("/roles", api.SetRole)           //create or update role
	router.Get("/levelpacks", api.GetLevelPacks) //get levelpacks
	router.Delete("/levelpack/:id", api.DeleteLevelPack)
	router.Post("/levelpack", api.EditLevelPack)

	router.Get("/get/users", api.QueryUsers)   //get users
	router.Get("/get/lusers", api.ListUsers)   //get chests
	router.Get("/get/levels", api.QueryLevels) //get levels

	router.Put("/modules/discord", api.DiscordModule)

	router.Get("/upgrade22", api.PressStart22Upgrade)
	router.Post("/buildlab", api.BuildLabPush)
	router.Get("/buildlab/status", api.GetBuildStatus)
	return nil
}

// Delete deletes gdps
// @Tags GDPS Management
// @Summary Deletes existing GDPS
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Param srvid path string true "GDPS Server ID"
// @Success 200 {object} structs.APIBasicSuccess
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /servers/gd/{srvid} [delete]
func (api *ManageGDPSAPI) Delete(c *fiber.Ctx) error {
	defer func() {
		if c := recover(); c != nil {
			debug.PrintStack()
		}
	}()
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	if err := utils.Should(srv.DeleteServer()); err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.NewAPIBasicResponse("Success"))
}

// Get returns gdps parameters
// @Tags GDPS Management
// @Summary Returns GDPS configuration
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Param srvid path string true "GDPS Server ID"
// @Success 200 {object} ServerGD.ServerGD
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /servers/gd/{srvid} [get]
func (api *ManageGDPSAPI) Get(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	_ = srv.LoadCoreConfig()
	srv.LoadTariff()
	return c.JSON(srv)
}

// ResetDBPassword resets gdps password
// @Tags GDPS Management
// @Summary Resets GDPS database password
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Param srvid path string true "GDPS Server ID"
// @Success 200 {object} structs.APIBasicSuccess
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /servers/gd/{srvid}/dbreset [get]
func (api *ManageGDPSAPI) ResetDBPassword(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	if err := utils.Should(srv.ResetDBPassword()); err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.NewAPIBasicResponse("Success"))
}

// UpdateChests updates gdps chest settings
// @Tags GDPS Management
// @Summary Updates GDPS chest settings
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Param srvid path string true "GDPS Server ID"
// @Param data body structs.ChestConfig true "All fields are required"
// @Success 200 {object} structs.APIBasicSuccess
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /servers/gd/{srvid}/chests [post]
func (api *ManageGDPSAPI) UpdateChests(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	var data structs.ChestConfig
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}

	if err := utils.Should(srv.UpdateChests(data)); err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}

	return c.JSON(structs.NewAPIBasicResponse("Success"))
}

// GetLogs  returns gdps logs by filter
// @Tags GDPS Management
// @Summary Retrieves GDPS logs by filter
// @Description -1=all, 0=registrations, 1=logins, 2=account deletions, 3=bans, 4=level actions
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Param srvid path string true "GDPS Server ID"
// @Param data body structs.APIManageGDLogsRequest true "All fields are required"
// @Success 200 {object} structs.APIManageGDLogsResponse
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /servers/gd/{srvid}/logs [post]
func (api *ManageGDPSAPI) GetLogs(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	var data structs.APIManageGDLogsRequest
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	logs, count, err := srv.GetLogs(data.Type, data.Page)
	if utils.Should(err) != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.APIManageGDLogsResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Count:           count,
		Results:         logs,
	})
}

// GetMusic  returns gdps music by filter
// @Tags GDPS Management
// @Summary Retrieves GDPS music by filter
// @Description Modes: id (asc), alpha (asc), downloads (desc)
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Param srvid path string true "GDPS Server ID"
// @Param data body structs.APIManageGDMusicRequest true "All fields are required"
// @Success 200 {object} structs.APIManageGDMusicResponse
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /servers/gd/{srvid}/music [post]
func (api *ManageGDPSAPI) GetMusic(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	var data structs.APIManageGDMusicRequest
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	songs, count, err := srv.SearchSongs(data.Query, data.Page, data.Mode)
	if utils.Should(err) != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.APIManageGDMusicResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Music:           songs,
		Count:           count,
	})
}

// AddMusic adds music to gdps
// @Tags GDPS Management
// @Summary Uploads music to GDPS
// @Description Types: ng (newgrounds), yt (youtube), vk (vkontakte), dz (deezer), db (dropbox/direct links)
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Param srvid path string true "GDPS Server ID"
// @Param data body structs.APIManageGDMusicAddRequest true "All fields are required"
// @Success 200 {object} structs.APIManageGDMusicAddResponse
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /servers/gd/{srvid}/music [put]
func (api *ManageGDPSAPI) AddMusic(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	var data structs.APIManageGDMusicAddRequest
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	music, err := srv.AddSong(data.Type, data.Url, "owner")
	if utils.Should(err) != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.APIManageGDMusicAddResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Music:           music,
	})
}

// ChangeIcon changes gdps icon
// @Tags GDPS Management
// @Summary Changes GDPS icon
// @Accept mpfd
// @Produce json
// @Param Authorization header string true "User token"
// @Param srvid path string true "GDPS Server ID"
// @Param profile_pic formData file true "Profile pic itself"
// @Success 200 {object} structs.APIBasicSuccess
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /servers/gd/{srvid}/icon [post]
func (api *ManageGDPSAPI) ChangeIcon(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	mpfd, err := c.MultipartForm()
	if utils.Should(err) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Unable to parse request", "form"))
	}
	icon := mpfd.File["profile_pic"]
	if len(icon) == 0 {
		return c.Status(500).JSON(structs.NewAPIError("No profile pic or reset flag provided", "file"))
	}
	iconf := icon[0]
	if iconf.Size > (5 << 20) {
		return c.Status(500).JSON(structs.NewAPIError("File is too big (>5MB)", "size"))
	}
	if !slices.Contains(fiberapi.ValidImageTypes, iconf.Header["Content-Type"][0]) {
		return c.Status(500).JSON(structs.NewAPIError("File is not an image", "type"))
	}
	iconImg, err := iconf.Open()
	var iconImage []byte
	if err == nil {
		iconImage, err = io.ReadAll(iconImg)
	}
	if err != nil {
		return c.Status(500).JSON(structs.NewAPIError("Error reading file", "file"))
	}
	if !slices.Contains(fiberapi.ValidImageTypes, http.DetectContentType(iconImage)) {
		return c.Status(500).JSON(structs.NewAPIError("File is not an image", "type"))
	}
	if err := srv.UpdateLogo(iconImage); err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.NewAPIBasicResponse("Success"))
}

// ChangeSettings changes gdps settings
// @Tags GDPS Management
// @Summary Changes GDPS settings
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Param srvid path string true "GDPS Server ID"
// @Param data body structs.GDSettings true "All fields are required"
// @Success 200 {object} structs.APIBasicSuccess
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /servers/gd/{srvid}/settings [post]
func (api *ManageGDPSAPI) ChangeSettings(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	var data structs.GDSettings
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	if err := srv.UpdateSettings(data); err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.NewAPIBasicResponse("Success"))
}

// BuildLabPush manages gdps installers
// @Tags GDPS Management
// @Summary Manages GDPS installers and mods
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Param srvid path string true "GDPS Server ID"
// @Param data body structs.BuildLabSettings true "All fields are required"
// @Success 200 {object} structs.APIBasicSuccess
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /servers/gd/{srvid}/buildlab [post]
func (api *ManageGDPSAPI) BuildLabPush(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	var data structs.BuildLabSettings
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	if err := srv.ExecuteBuildLab(data); err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.NewAPIBasicResponse("Success"))
}

func (api *ManageGDPSAPI) GetBuildStatus(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	return c.JSON(structs.NewAPIBasicResponse(srv.FetchBuildStatus()))
}

func (api *ManageGDPSAPI) DiscordModule(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	body := c.Request().Body()
	var data interface{}
	_ = json.Unmarshal(body, &data)
	pdata := data.(map[string]interface{})
	err := srv.ModuleDiscord(pdata["enable"].(bool), pdata)
	if err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.NewAPIBasicResponse("Success"))
}

func (api *ManageGDPSAPI) GetRoles(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	interactor := srv.NewInteractor()
	defer interactor.Dispose()
	roles := interactor.GetRoles()
	return c.JSON(structs.APIRolesResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Roles:           roles,
	})
}

func (api *ManageGDPSAPI) SetRole(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}

	var data structs.InjectedGDRole
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	interactor := srv.NewInteractor()
	defer interactor.Dispose()
	err := interactor.SetRole(data)
	if err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.NewAPIBasicResponse("Success"))
}

func (api *ManageGDPSAPI) QueryUsers(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	interactor := srv.NewInteractor()
	defer interactor.Dispose()
	users := interactor.SearchUsers(c.Query("user"))
	return c.JSON(structs.APIGDPSUsersResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Users:           users,
	})
}
func (api *ManageGDPSAPI) ListUsers(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	interactor := srv.NewInteractor()
	defer interactor.Dispose()
	users := interactor.ListUsers(c.QueryInt("page"))
	return c.JSON(structs.APIGDPSUsersResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Users:           users,
	})
}

func (api *ManageGDPSAPI) GetLevelPacks(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	interactor := srv.NewInteractor()
	defer interactor.Dispose()
	packs := interactor.GetPacks(c.QueryBool("gau"))
	return c.JSON(structs.APILevelpacksResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Packs:           packs,
	})
}

func (api *ManageGDPSAPI) DeleteLevelPack(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	param, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	interactor := srv.NewInteractor()
	defer interactor.Dispose()
	err = interactor.DeletePack(param)
	if err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.NewAPIBasicResponse("Success"))
}

func (api *ManageGDPSAPI) EditLevelPack(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}

	var data structs.InjectedGDLevelPack
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	interactor := srv.NewInteractor()
	defer interactor.Dispose()
	err := interactor.SetPack(data)
	if err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.NewAPIBasicResponse("Success"))
}

func (api *ManageGDPSAPI) QueryLevels(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	interactor := srv.NewInteractor()
	defer interactor.Dispose()
	levels := interactor.SearchLevels(c.Query("lvl"))
	return c.JSON(structs.APIGDPSLevelsResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Users:           levels,
	})
}

func (api *ManageGDPSAPI) PressStart22Upgrade(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.PerformAuth_(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	if err := srv.ExecuteBuildLab(structs.BuildLabSettings{
		Version:  "2.2",
		Windows:  true,
		Android:  true,
		Textures: "default",
	}); err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.NewAPIBasicResponse("Success"))
}

// -------------------
// authGDPS is a simple authenticator for compacting code
func (api *ManageGDPSAPI) authGDPS(c *fiber.Ctx, acc *providers.Account, srv *ServerGD.ServerGD) bool {
	srvid := c.Params("srvid")
	if len(srvid) != 4 || !srv.GetServerBySrvID(srvid) {
		return false
	}
	if srv.Srv.OwnerID != acc.Data().UID && !acc.Data().IsAdmin {
		// If not owner and not admin, then GG
		return false
	}
	return true
}
