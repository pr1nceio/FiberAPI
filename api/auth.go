package api

import (
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/api/ent"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/gofiber/fiber/v2"
	"strings"
)

type AuthAPI struct {
	*ent.API
}

func (api *AuthAPI) Register(router fiber.Router) error {
	router.Post("/login", api.Login)
	router.Post("/register", api.ARegister)
	router.All("/confirm_email", api.ConfirmEmail)
	router.Post("/recover", api.RecoverPassword)
	router.All("/discord", api.Discord)
	return nil
}

// AuthRegister registers new FruitSpace user
// @Tags Authentication
// @Summary Registers new FruitSpace user
// @Accept json
// @Produce json
// @Param data body structs.AuthRegisterRequest true "All fields are required"
// @Param Cookie header string false "affiliate=code"
// @Success 200 {object} structs.APIBasicSuccess
// @Failure 500 {object} structs.APIError
// @Router /auth/register [post]
func (api *AuthAPI) ARegister(c *fiber.Ctx) error {
	var data structs.AuthRegisterRequest
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	if !utils.VerifyCaptcha(data.HCaptchaToken, fiberapi.CONFIG["hCaptchaToken"]) {
		return c.Status(500).JSON(structs.NewAPIError("Captcha failed", "captcha"))
	}
	acc := api.AccountProvider.New()
	if err := acc.Register(data.Uname, data.Name, data.Surname, data.Email, data.Password, c.Cookies("affiliate"), getIP(c), data.Lang); err != nil {
		return c.Status(500).JSON(structs.NewDecoupleAPIError(err))
	}
	return c.JSON(structs.NewAPIBasicResponse("Account created"))
}

// AuthLogin logs in new FruitSpace user
// @Tags Authentication
// @Summary Logs in by provided credentials and returns session
// @Accept json
// @Produce json
// @Param data body structs.AuthLoginRequest true "All fields are required"
// @Success 200 {object} structs.AuthLoginResponse
// @Failure 500 {object} structs.APIError
// @Router /auth/login [post]
func (api *AuthAPI) Login(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	var data structs.AuthLoginRequest
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	if err := acc.Login(data.Uname, data.Password, getIP(c)); err != nil {
		return c.Status(500).JSON(structs.NewDecoupleAPIError(err))
	}
	if acc.Data().Is2FA {
		if len(data.TOTP) == 0 {
			return c.JSON(structs.NewAPIError("TOTP code is required", "2fa_req"))
		}
		secret, _ := acc.CreateTOTP(data.TOTP)
		if len(secret) == 0 {
			return c.Status(500).JSON(structs.NewAPIError("Invalid 2FA Code", "2fa"))
		}
	}
	if !utils.VerifyCaptcha(data.HCaptchaToken, fiberapi.CONFIG["hCaptchaToken"]) {
		return c.Status(500).JSON(structs.NewAPIError("Captcha failed", "captcha"))
	}
	token := acc.NewSession(acc.Data().UID, getUserAgent(c), getIP(c))
	api.SetToken_(c, token)
	return c.JSON(structs.AuthLoginResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Logged in"),
		Token:           token,
	})
}

// AuthConfirmEmail serves page for email confirmations [/auth/confirm_email]
func (api *AuthAPI) ConfirmEmail(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	uid := acc.DecodeEmailToken(c.Query("token"))
	if !acc.GetUserByUID(uid) {
		return c.Status(500).JSON(structs.NewAPIError("Invalid token supplied"))
	}
	if err := acc.VerifyEmail(); err != nil {
		return c.Status(500).JSON(structs.NewAPIError("Unable to activate account", err.Error()))
	}
	r, _ := fiberapi.AssetsDir.ReadFile("assets/EmailConfirmationIndex.html")
	c.Set("Content-Type", "text/html")
	token := acc.NewSession(acc.Data().UID, getUserAgent(c), getIP(c))
	api.SetToken_(c, token)
	return c.SendString(strings.ReplaceAll(strings.ReplaceAll(string(r),
		"{uname}", acc.Data().Uname), "{token}", token,
	))
}

// AuthRecoverPassword sends email to user with randomly generated password
// @Tags Authentication
// @Summary Sends email with randomly generated password
// @Accept json
// @Produce json
// @Param data body structs.AuthRecoverRequest true "All fields are required"
// @Success 200 {object} structs.APIBasicSuccess
// @Failure 500 {object} structs.APIError
// @Router /auth/recover [post]
func (api *AuthAPI) RecoverPassword(c *fiber.Ctx) error {
	var data structs.AuthRecoverRequest
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}
	if !utils.VerifyCaptcha(data.HCaptchaToken, fiberapi.CONFIG["hCaptchaToken"]) {
		return c.Status(500).JSON(structs.NewAPIError("Captcha failed", "captcha"))
	}
	acc := api.AccountProvider.New()
	if err := acc.RecoverPassword(data.Email, data.Lang); err != nil {
		return c.Status(500).JSON(structs.NewDecoupleAPIError(err))
	}
	return c.JSON(structs.NewAPIBasicResponse("New password sent to your email inbox"))
}

func (api *AuthAPI) Discord(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	err := acc.AuthDiscord(c.Query("code"), c.Query("state"))
	if utils.Should(err) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Unauthorized"))
	}
	token := acc.NewSession(acc.Data().UID, getUserAgent(c), getIP(c))
	r, _ := fiberapi.AssetsDir.ReadFile("assets/DiscordConfirmationIndex.html")
	c.Set("Content-Type", "text/html")
	api.SetToken_(c, token)
	return c.SendString(strings.ReplaceAll(strings.ReplaceAll(string(r), "{uname}", acc.Data().Uname), "{token}", token))
}
