package api

import (
	"fmt"
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/models/gdps_db"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/providers"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/gofiber/fiber/v2"
	"strconv"
	"strings"
)

func (api *API) AuxiliaryGDPSLogin(c *fiber.Ctx) error {
	srvid := c.Params("srvid")
	var data structs.AuthLoginRequest
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	if !utils.VerifyCaptcha(data.HCaptchaToken, fiberapi.CONFIG["hCaptchaToken"]) {
		return c.Status(500).JSON(structs.NewAPIError("Captcha failed", "captcha"))
	}
	srv := api.ServerGDProvider.New()
	if len(srvid) != 4 || !srv.Exists(srvid) {
		return c.Status(500).JSON(structs.NewAPIError("No server found"))
	}
	acc := srv.NewGDPSUser()
	defer acc.Dispose()
	res := acc.LogIn(data.Uname, data.Password, getIP(c), 0, false)
	if res > 0 {
		token := fmt.Sprintf("%d:%s", acc.Data().UID, acc.Data().Passhash)
		return c.JSON(structs.AuthLoginResponse{
			APIBasicSuccess: structs.NewAPIBasicResponse("Logged in"),
			Token:           token,
		})
	}
	return c.JSON(structs.NewAPIError("Error", strconv.Itoa(res)))
}

func (api *API) AuxiliaryGDPSAuth(c *fiber.Ctx) error {
	srv := api.ServerGDProvider.New()
	var acc *providers.ServerGDUser
	defer acc.Dispose()
	err := api.gdpsUserAuth(c, acc, srv)
	if err != nil {
		return err
	}
	return c.JSON(struct {
		structs.APIBasicSuccess
		*gdps_db.User
	}{
		structs.NewAPIBasicResponse("Success"),
		acc.Data(),
	})
}

func (api *API) AuxiliaryGDPSAddMusic(c *fiber.Ctx) error {
	srv := api.ServerGDProvider.New()
	var acc *providers.ServerGDUser
	defer acc.Dispose()
	err := api.gdpsUserAuth(c, acc, srv)
	if err != nil {
		return err
	}

	var data structs.APIManageGDMusicAddRequest
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	music, err := srv.AddSong(data.Type, data.Url)
	if utils.Should(err) != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.APIManageGDMusicAddResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Music:           music,
	})
}

func (api *API) AuxiliaryGDPSGetMusic(c *fiber.Ctx) error {
	srv := api.ServerGDProvider.New()
	var acc *providers.ServerGDUser
	defer acc.Dispose()
	err := api.gdpsUserAuth(c, acc, srv)
	if err != nil {
		return err
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

func (api *API) gdpsUserAuth(c *fiber.Ctx, acc *providers.ServerGDUser, srv *providers.ServerGD) error {
	srvid := c.Params("srvid")
	token := strings.Split(c.Get("Authorization"), ":")
	if len(token) != 2 {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	uid, _ := strconv.Atoi(token[0])
	if len(srvid) != 4 || !srv.Exists(srvid) {
		return c.Status(500).JSON(structs.NewAPIError("No server found"))
	}
	acc = srv.NewGDPSUser()
	res := acc.LogIn("", token[1], getIP(c), uid, true)
	if res > 0 {
		return nil
	}
	return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
}
