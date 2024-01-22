package api

import (
	"errors"
	"fmt"
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/models/gdps_db"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/providers/ServerGD"
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
	val := false
	if len(data.FCaptchaToken) > 0 {
		val = utils.VerifyFCaptcha(data.FCaptchaToken, fiberapi.CONFIG["fCaptchaToken"])
	}
	//else if len(data.HCaptchaToken) > 0 {
	//	val = utils.VerifyCaptcha(data.HCaptchaToken, fiberapi.CONFIG["hCaptchaToken"])
	//}
	if !val {
		return c.Status(500).JSON(structs.NewAPIError("Captcha failed", "captcha"))
	}
	srv := api.ServerGDProvider.New()
	if len(srvid) != 4 || !srv.Exists(srvid) {
		return c.Status(500).JSON(structs.NewAPIError("No server found"))
	}
	acc := srv.NewGDPSUser()
	defer func() {
		if acc != nil {
			acc.Dispose()
		}
	}()
	res := acc.LogIn(data.Uname, data.Password, getIP(c), 0, false)
	if res > 0 {
		srv.LoadCoreConfig()
		if acc.Data().IsBanned == 1 && srv.CoreConfig.ServerConfig.EnableModules["discord"] {
			srv.SendWebhook("newuser", map[string]string{"nickname": data.Uname})
		}

		token := fmt.Sprintf("%d:%s", acc.Data().UID, acc.Data().Passhash)
		return c.JSON(structs.AuthLoginResponse{
			APIBasicSuccess: structs.NewAPIBasicResponse("Logged in"),
			Token:           token,
		})
	}
	return c.JSON(structs.NewAPIError("Error", strconv.Itoa(res)))
}

func (api *API) AuxiliaryGDPSForgotPassword(c *fiber.Ctx) error {
	srvid := c.Params("srvid")
	var data struct {
		Email         string `json:"email"`
		HCaptchaToken string `json:"hCaptchaToken"`
		FCaptchaToken string `json:"fCaptchaToken"`
	}
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	val := false
	if len(data.FCaptchaToken) > 0 {
		val = utils.VerifyFCaptcha(data.FCaptchaToken, fiberapi.CONFIG["fCaptchaToken"])
	}
	//else if len(data.HCaptchaToken) > 0 {
	//	val = utils.VerifyCaptcha(data.HCaptchaToken, fiberapi.CONFIG["hCaptchaToken"])
	//}
	if !val {
		return c.Status(500).JSON(structs.NewAPIError("Captcha failed", "captcha"))
	}
	srv := api.ServerGDProvider.New()
	if len(srvid) != 4 || !srv.Exists(srvid) {
		return c.Status(500).JSON(structs.NewAPIError("No server found"))
	}
	acc := srv.NewGDPSUser()
	defer func() {
		if acc != nil {
			acc.Dispose()
		}
	}()
	if !acc.GetUserByEmail(data.Email) {
		return c.JSON(structs.NewAPIError("No user found", "nouser"))
	}
	err := acc.UserForgotPasswordSendEmail(srvid)
	if err == nil {
		return c.JSON(structs.NewAPIBasicResponse("Success"))
	}
	return c.JSON(structs.NewDecoupleAPIError(err))
}

func (api *API) AuxiliaryGDPSAuth(c *fiber.Ctx) error {
	srv := api.ServerGDProvider.New()
	acc, err := api.gdpsUserAuth(c, srv)
	defer func() {
		if acc != nil {
			acc.Dispose()
		}
	}()
	if acc == nil {
		return c.JSON(structs.NewAPIError(err.Error()))
	}
	data := struct {
		structs.APIBasicSuccess
		*gdps_db.User
	}{
		structs.NewAPIBasicResponse("Success"),
		acc.Data(),
	}
	return c.JSON(data)
}

func (api *API) AuxiliaryGDPSChangeCreds(c *fiber.Ctx) error {
	srv := api.ServerGDProvider.New()
	var data struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"uname"`
	}
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	acc, err := api.gdpsUserAuth(c, srv)
	defer func() {
		if acc != nil {
			acc.Dispose()
		}
	}()
	if err != nil {
		return c.JSON(structs.NewAPIError(err.Error()))
	}
	if data.Email != "" {
		err = acc.UserChangeEmail(data.Email)
	}
	if data.Username != "" {
		err = acc.UserChangeUsername(data.Username)
	}
	if data.Password != "" {
		err = acc.UserChangePassword(data.Password)
	}
	if utils.Should(err) == nil {
		return c.JSON(structs.NewAPIBasicResponse("Success"))
	}
	return c.JSON(structs.NewDecoupleAPIError(err))
}

func (api *API) AuxiliaryGDPSAddMusic(c *fiber.Ctx) error {
	srv := api.ServerGDProvider.New()
	acc, err := api.gdpsUserAuth(c, srv)
	defer func() {
		if acc != nil {
			acc.Dispose()
		}
	}()
	if err != nil {
		return c.JSON(structs.NewAPIError(err.Error()))
	}

	var data structs.APIManageGDMusicAddRequest
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	music, err := srv.AddSong(data.Type, data.Url, strconv.Itoa(acc.Data().UID)+"_"+acc.Data().Uname)
	if utils.Should(err) != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	srv.LoadCoreConfig()
	if srv.CoreConfig.ServerConfig.EnableModules["discord"] {
		srv.SendWebhook("newmusic", map[string]string{
			"id":       strconv.Itoa(music.ID),
			"name":     music.Name,
			"artist":   music.Artist,
			"nickname": acc.Data().Uname,
		})
	}
	return c.JSON(structs.APIManageGDMusicAddResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Music:           music,
	})
}

func (api *API) AuxiliaryGDPSGetMusic(c *fiber.Ctx) error {
	srv := api.ServerGDProvider.New()
	acc, err := api.gdpsUserAuth(c, srv)
	defer func() {
		if acc != nil {
			acc.Dispose()
		}
	}()
	if err != nil {
		return c.JSON(structs.NewAPIError(err.Error()))
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

func (api *API) AuxiliaryGDPSChangePassword(c *fiber.Ctx) error {
	srv := api.ServerGDProvider.New()
	acc, err := api.gdpsUserAuth(c, srv)
	defer func() {
		if acc != nil {
			acc.Dispose()
		}
	}()
	if err != nil {
		return c.JSON(structs.NewAPIError(err.Error()))
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

func (api *API) gdpsUserAuth(c *fiber.Ctx, srv *ServerGD.ServerGD) (*ServerGD.ServerGDUser, error) {
	srvid := c.Params("srvid")
	token := strings.Split(c.Get("Authorization"), ":")
	if len(token) != 2 {
		return nil, errors.New("Unauthorized")
	}
	uid, _ := strconv.Atoi(token[0])
	if len(srvid) != 4 || !srv.Exists(srvid) {
		return nil, errors.New("No server found")
	}
	acc := srv.NewGDPSUser()
	res := acc.LogIn("", token[1], getIP(c), uid, true)
	if res > 0 {
		return acc, nil
	}
	return nil, errors.New("Unauthorized")
}
