package api

import (
	"fmt"
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/providers"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/slices"
	"io"
	"net/http"
)

// UserSSO returns base information
// @Tags User Management
// @Summary Returns base user information
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Success 200 {object} structs.APIUserSSO
// @Failure 403 {object} structs.APIError
// @Failure 500 {object} structs.APIError
// @Router /user [get]
func (api *API) UserSSO(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	if !api.performAuth(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	aData := acc.Data()
	return c.JSON(structs.APIUserSSO{
		Status:        "ok",
		Uname:         aData.Uname,
		Name:          aData.Name,
		Surname:       aData.Surname,
		ProfilePic:    fmt.Sprintf("https://%s/profile_pics/%s", fiberapi.S3_CONFIG["cdn"], aData.ProfilePic),
		VkId:          aData.VkID,
		DiscordId:     aData.DiscordID,
		Balance:       aData.Balance,
		ShopBalance:   api.ShopProvider.GetUserShopsBalance(aData.UID),
		Is2FA:         aData.Is2FA,
		IsAdmin:       aData.IsAdmin,
		IsPartner:     aData.IsPartner,
		Reflink:       aData.Reflink,
		Notifications: api.NotificationProvider.GetNotificationsForUID(aData.UID),
		Servers:       acc.GetServersCount(),
		TopServers: map[string]interface{}{
			"gd": api.ServerGDProvider.New().GetTopUserServer(aData.UID),
		},
	})
}

// UserUpdate updates name, password and 2FA
// @Tags User Management
// @Summary Update username, password and 2FA
// @Accept json
// @Produce json
// @Param Authorization header string true "User token"
// @Param data body structs.APIUserUpdateRequest true "Only non-empty fields are updated"
// @Success 200 {object} structs.UserUpdateResponse
// @Failure 403 {object} structs.APIError
// @Failure 500 {object} structs.APIError
// @Router /user [patch]
func (api *API) UserUpdate(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	if !api.performAuth(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	var data structs.APIUserUpdateRequest
	if c.BodyParser(&data) != nil {
		return c.Status(500).JSON(structs.NewAPIError("Invalid request"))
	}

	var secret, img string

	// Update Name+Surname
	if len(data.Name) > 0 && len(data.Surname) > 0 {
		err := acc.UpdateNameSurname(data.Name, data.Surname)
		if err != nil {
			return c.Status(500).JSON(structs.NewDecoupleAPIError(err))
		}
	}

	// Update Password
	if len(data.Password) > 0 && len(data.NewPassword) > 0 {
		err := acc.UpdatePassword(data.Password, data.NewPassword)
		if err != nil {
			return c.Status(500).JSON(structs.NewDecoupleAPIError(err))
		}
	}

	// 2FA
	if len(data.TOTP) > 0 {
		secret, img = acc.CreateTOTP(data.TOTP)
		if len(secret) == 0 {
			return c.Status(500).JSON(structs.NewAPIError("2FA is already enabled", "2fa_enabled"))
		}
	}

	return c.JSON(structs.UserUpdateResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("No errors occurred"),
		TotpSecret:      secret,
		TotpImage:       img,
	})
}

// UserAvatarUpdate updates user profile pic
// @Tags User Management
// @Summary Update user profile pic
// @Accept mpfd
// @Produce json
// @Param Authorization header string true "User token"
// @Param reset formData string false "Should profile pic be reset"
// @Param profile_pic formData file true "Profile pic itself"
// @Success 200 {object} structs.UserAvatarResponse
// @Failure 403 {object} structs.APIError
// @Failure 500 {object} structs.APIError
// @Router /user/avatar [post]
func (api *API) UserAvatarUpdate(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	if !api.performAuth(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	mpfd, err := c.MultipartForm()
	if err != nil {
		return c.Status(500).JSON(structs.NewAPIError("Unable to parse request", "form"))
	}
	if len(mpfd.Value["reset"]) > 0 && mpfd.Value["reset"][0] == "reset" {
		acc.ResetProfilePic()
		return c.JSON(structs.UserAvatarResponse{
			APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
			ProfilePic:      fmt.Sprintf("https://%s/profile_pics/%s", fiberapi.S3_CONFIG["cdn"], acc.Data().ProfilePic),
		})
	}

	pfp := mpfd.File["profile_pic"]
	if len(pfp) == 0 {
		return c.Status(500).JSON(structs.NewAPIError("No profile pic or reset flag provided", "file"))
	}
	pfpf := pfp[0]
	if pfpf.Size > (5 << 20) {
		return c.Status(500).JSON(structs.NewAPIError("File is too big (>5MB)", "size"))
	}
	if !slices.Contains(fiberapi.ValidImageTypes, pfpf.Header["Content-Type"][0]) {
		return c.Status(500).JSON(structs.NewAPIError("File is not an image", "type"))
	}
	pfImg, err := pfpf.Open()
	var pfpImage []byte
	if err == nil {
		pfpImage, err = io.ReadAll(pfImg)
	}
	if err != nil {
		return c.Status(500).JSON(structs.NewAPIError("Error reading file", "file"))
	}
	if !slices.Contains(fiberapi.ValidImageTypes, http.DetectContentType(pfpImage)) {
		return c.Status(500).JSON(structs.NewAPIError("File is not an image", "type"))
	}
	if err = acc.UpdateProfilePic(pfpImage); err != nil {
		return c.Status(500).JSON(structs.NewAPIError(err.Error()))
	}
	return c.JSON(structs.UserAvatarResponse{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		ProfilePic:      fmt.Sprintf("https://%s/profile_pics/%s", fiberapi.S3_CONFIG["cdn"], acc.Data().ProfilePic),
	})
}

func (api *API) UserJoinGuild(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	if !api.performAuth(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	err := acc.DiscordJoinGuild()
	if err != nil {
		return c.JSON(structs.NewAPIError("Couldn't autoinvite"))
	}
	return c.JSON(structs.NewAPIBasicResponse("Success"))
}

func (api *API) UserListSessions(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	if !api.performAuth(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	return c.JSON(struct {
		structs.APIBasicSuccess
		Sessions []*providers.Session `json:"sessions"`
	}{
		APIBasicSuccess: structs.NewAPIBasicResponse("Success"),
		Sessions:        acc.ListSessions(),
	})
}

func (api *API) UserDeleteSession(c *fiber.Ctx) error {
	acc := api.AccountProvider.New()
	if !api.performAuth(c, acc) {
		return c.Status(403).JSON(structs.NewAPIError("Unauthorized"))
	}
	sess := c.Params("session")
	acc.DeleteSession(sess)
	return c.JSON(structs.NewAPIBasicResponse("Success"))
}
