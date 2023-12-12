package api

import (
	"encoding/json"
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/providers"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/slices"
	"io"
	"net/http"
	"runtime/debug"
)

// ManageGDPSDelete deletes gdps
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
func (api *API) ManageGDPSDelete(c *fiber.Ctx) error {
	defer func() {
		if c := recover(); c != nil {
			debug.PrintStack()
		}
	}()
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
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

// ManageGDPSGet returns gdps parameters
// @Tags GDPS Management
// @Summary Returns GDPS configuration
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Param srvid path string true "GDPS Server ID"
// @Success 200 {object} providers.ServerGD
// @Failure 500 {object} structs.APIError
// @Failure 403 {object} structs.APIError
// @Router /servers/gd/{srvid} [get]
func (api *API) ManageGDPSGet(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	srv.LoadCoreConfig()
	srv.LoadTariff()
	return c.JSON(srv)
}

// ManageGDPSResetDBPassword resets gdps password
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
func (api *API) ManageGDPSResetDBPassword(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
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

// ManageGDPSUpdateChests updates gdps chest settings
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
func (api *API) ManageGDPSUpdateChests(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
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

// ManageGDPSGetLogs  returns gdps logs by filter
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
func (api *API) ManageGDPSGetLogs(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
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

// ManageGDPSGetMusic  returns gdps music by filter
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
func (api *API) ManageGDPSGetMusic(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
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

// ManageGDPSAddMusic adds music to gdps
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
func (api *API) ManageGDPSAddMusic(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
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

// ManageGDPSChangeIcon changes gdps icon
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
func (api *API) ManageGDPSChangeIcon(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
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

// ManageGDPSChangeSettings changes gdps settings
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
func (api *API) ManageGDPSChangeSettings(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
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

// ManageGDPSBuildLabPush manages gdps installers
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
func (api *API) ManageGDPSBuildLabPush(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
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

func (api *API) ManageGDPSGetBuildStatus(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	return c.JSON(structs.NewAPIBasicResponse(srv.FetchBuildStatus()))
}

func (api *API) ManageGDPSDiscordModule(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	if !api.authGDPS(c, acc, srv) {
		return c.Status(500).JSON(structs.NewAPIError("You have no permission to manage this server"))
	}
	body := c.Request().Body()
	var data interface{}
	json.Unmarshal(body, &data)
	pdata := data.(map[string]interface{})
	err := srv.ModuleDiscord(pdata["enable"].(bool), pdata)
	if err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.NewAPIBasicResponse("Success"))
}

func (api *API) ManageGDPSGetRoles(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
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

func (api *API) ManageGDPSSetRole(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
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

func (api *API) ManageGDPSQueryUsers(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	srv := api.ServerGDProvider.New()
	if !api.performAuth(c, acc) {
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

// -------------------
// authGDPS is a simple authenticator for compacting code
func (api *API) authGDPS(c *fiber.Ctx, acc *providers.Account, srv *providers.ServerGD) bool {
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
