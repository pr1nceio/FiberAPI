package structs

import (
	"encoding/json"
	"github.com/fruitspace/FiberAPI/models/db"
	"strings"
)

type MusicResponse struct {
	Status string      `json:"status"`
	Name   string      `json:"name"`
	Artist string      `json:"artist"`
	Size   json.Number `json:"size"`
	Url    string      `json:"url"`
}

type PaymentResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Url     string `json:"url"`
}

type AuthRegisterRequest struct {
	Uname         string `json:"uname"`
	Name          string `json:"name"`
	Surname       string `json:"surname"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	Lang          string `json:"lang"`
	HCaptchaToken string `json:"hCaptchaToken"`
}

type AuthLoginRequest struct {
	Uname         string `json:"uname"`
	Password      string `json:"password"`
	HCaptchaToken string `json:"hCaptchaToken"`
}

type AuthRecoverRequest struct {
	Email         string `json:"email"`
	HCaptchaToken string `json:"hCaptchaToken"`
	Lang          string `json:"lang"`
}

type APIError struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewAPIError(err string, code ...string) APIError {
	pcode := "generic"
	if len(code) > 0 {
		pcode = code[0]
	}
	return APIError{Status: "error", Message: err, Code: pcode}
}

func NewDecoupleAPIError(err error) APIError {
	dc := strings.Split(err.Error(), "|")
	return NewAPIError(dc[0], dc[1:]...)
}

type APIBasicSuccess struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewAPIBasicResponse(msg string) APIBasicSuccess {
	return APIBasicSuccess{Status: "ok", Message: msg}
}

type AuthLoginResponse struct {
	APIBasicSuccess
	Token string `json:"token"`
}

type UserUpdateResponse struct {
	APIBasicSuccess
	TotpSecret string `json:"totp_secret"`
	TotpImage  string `json:"totp_image"`
}

type UserAvatarResponse struct {
	APIBasicSuccess
	ProfilePic string `json:"profile_pic"`
}

type APIUserSSO struct {
	Status        string                 `json:"status"`
	Uname         string                 `json:"uname"`
	Name          string                 `json:"name"`
	Surname       string                 `json:"surname"`
	ProfilePic    string                 `json:"profile_pic"`
	Bal           float64                `json:"bal"`
	ShopBal       float64                `json:"shop_bal"`
	Is2FA         bool                   `json:"is2fa"`
	IsAdmin       bool                   `json:"is_admin"`
	Reflink       string                 `json:"reflink"`
	Notifications []db.Notification      `json:"notifications"`
	Servers       map[string]int         `json:"servers"`
	TopServers    map[string]interface{} `json:"top_servers"`
}

type APIUserUpdateRequest struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
	TOTP        string `json:"totp"`
}

type APIPaymentRequest struct {
	Amount   string `json:"amount"`
	Merchant string `json:"merchant"`
}
