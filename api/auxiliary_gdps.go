package api

import (
	"errors"
	"fmt"
	fiberapi "github.com/fruitspace/HyprrSpace"
	"github.com/fruitspace/HyprrSpace/api/ent"
	"github.com/fruitspace/HyprrSpace/models/gdps_db"
	"github.com/fruitspace/HyprrSpace/models/structs"
	"github.com/fruitspace/HyprrSpace/providers/ServerGD"
	"github.com/fruitspace/HyprrSpace/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"strconv"
	"strings"
)

type AuxiliaryGDPSAPI struct {
	*ent.API
}

func (api *AuxiliaryGDPSAPI) Register(router fiber.Router) error {
	router.Post("/login", api.Login)
	router.Get("/", api.Auth)
	router.Put("/", api.ChangeCreds)
	router.Post("/music", api.GetMusic)
	router.Put("/music", api.AddMusic)
	router.Post("/recover", api.ForgotPassword)
	router.Group("/gdproxy").Use(func(c *fiber.Ctx) error {
		c.Request().Header.Set("User-Agent", "")
		c.Request().SetHost("www.boomlings.com")
		return proxy.Do(c, "https://www.boomlings.com/database"+strings.Split(c.Path(), "/gdproxy")[1])
	})
	return nil
}

func (api *AuxiliaryGDPSAPI) Login(c *fiber.Ctx) error {
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

func (api *AuxiliaryGDPSAPI) ForgotPassword(c *fiber.Ctx) error {
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

func (api *AuxiliaryGDPSAPI) Auth(c *fiber.Ctx) error {
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

func (api *AuxiliaryGDPSAPI) ChangeCreds(c *fiber.Ctx) error {
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

func (api *AuxiliaryGDPSAPI) AddMusic(c *fiber.Ctx) error {
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

func (api *AuxiliaryGDPSAPI) GetMusic(c *fiber.Ctx) error {
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

func (api *AuxiliaryGDPSAPI) ChangePassword(c *fiber.Ctx) error {
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

func (api *AuxiliaryGDPSAPI) gdpsUserAuth(c *fiber.Ctx, srv *ServerGD.ServerGD) (*ServerGD.ServerGDUser, error) {
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
