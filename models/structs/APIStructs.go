package structs

import (
	"encoding/json"
	"github.com/fruitspace/FiberAPI/models/db"
	"github.com/fruitspace/FiberAPI/models/gdps_db"
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
	TOTP          string `json:"totp"`
	HCaptchaToken string `json:"hCaptchaToken"`
	FCaptchaToken string `json:"fCaptchaToken"`
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
	return NewAPIError(strings.TrimSpace(dc[0]), dc[1:]...)
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
	Status      string                 `json:"status"`
	Uname       string                 `json:"uname"`
	Name        string                 `json:"name"`
	Surname     string                 `json:"surname"`
	ProfilePic  string                 `json:"profile_pic"`
	VkId        string                 `json:"vk_id"`
	DiscordId   string                 `json:"discord_id"`
	Balance     float64                `json:"balance"`
	ShopBalance float64                `json:"shop_balance"`
	Is2FA       bool                   `json:"is_2fa"`
	IsAdmin     bool                   `json:"is_admin"`
	IsPartner   bool                   `json:"is_partner"`
	Reflink     string                 `json:"reflink"`
	Servers     map[string]int64       `json:"servers"`
	TopServers  map[string]interface{} `json:"top_servers"`
}

type APIUserUpdateRequest struct {
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
	TOTP        string `json:"totp"`
}

type APIPaymentRequest struct {
	Amount   float64 `json:"amount"`
	Merchant string  `json:"merchant"`
}

type APIPaymentResponse struct {
	APIBasicSuccess
	Url string `json:"pay_url"`
}

type APIPaymentListResponse struct {
	APIBasicSuccess
	Transactions []*db.Transaction `json:"transactions"`
}

type APIServerListResponse struct {
	APIBasicSuccess
	GD []*db.ServerGdSmall `json:"gd"`
	MC []*db.ServerMc      `json:"mc"`
	CS []string            `json:"cs"`
}

type APIServerGDCreateRequest struct {
	Name      string `json:"name"`
	SrvId     string `json:"srvid"`
	Tariff    int    `json:"tariff"`
	Duration  string `json:"duration"`
	Promocode string `json:"promocode"`
}

type APIManageGDLogsRequest struct {
	Page int `json:"page"`
	Type int `json:"type"`
}

type APIManageGDLogsResponse struct {
	APIBasicSuccess
	Count   int               `json:"count"`
	Results []*gdps_db.Action `json:"results"`
}

type APIManageGDMusicRequest struct {
	Query string `json:"query"`
	Page  int    `json:"page"`
	Mode  string `json:"mode"`
}

type APIManageGDMusicResponse struct {
	APIBasicSuccess
	Music []*gdps_db.Song `json:"music"`
	Count int             `json:"count"`
}

type APIManageGDMusicAddRequest struct {
	Url  string `json:"url"`
	Type string `json:"type"`
}

type APIManageGDMusicAddResponse struct {
	APIBasicSuccess
	Music *gdps_db.Song `json:"music"`
}

type APIFetchDiscordUser struct {
	APIBasicSuccess
	UID     int                 `json:"uid"`
	Uname   string              `json:"uname"`
	Avatar  string              `json:"avatar"`
	Active  string              `json:"active"`
	Balance int                 `json:"balance"`
	Servers []*db.ServerGdSmall `json:"servers"`
}

type APITopGDServers struct {
	APIBasicSuccess
	Servers []*db.ServerGdSmall `json:"servers"`
}

type InjectedGDRole struct {
	gdps_db.Role
	Users []gdps_db.UserNano `json:"users" gorm:"-"`
}

type InjectedGDLevelPack struct {
	gdps_db.LevelPack
	Levels []gdps_db.LevelNano `json:"levels" gorm:"-"`
}

type InjectedRequestGDLevelPack struct {
	gdps_db.LevelPack
	Levels []int `json:"levels" gorm:"-"`
}

type APIRolesResponse struct {
	APIBasicSuccess
	Roles []InjectedGDRole `json:"roles"`
}

type APIGDPSUsersResponse struct {
	APIBasicSuccess
	Users []gdps_db.UserNano `json:"users"`
	Count int64              `json:"count"`
}

type APIGDPSLevelsResponse struct {
	APIBasicSuccess
	Users []gdps_db.LevelNano `json:"levels"`
}

type APILevelpacksResponse struct {
	APIBasicSuccess
	Packs []InjectedGDLevelPack `json:"packs"`
}

type APIParticleSearchRequest struct {
	Query      string   `json:"query"`
	Arch       []string `json:"arch"`
	IsOfficial bool     `json:"is_official"`
	Sort       string   `json:"sort"`
	Page       int      `json:"page"`
}

type APIParticleSearchResponse struct {
	APIBasicSuccess
	Particles []db.Particle `json:"particles"`
	Count     int64         `json:"count"`
}

type APIParticleUserResponse struct {
	APIBasicSuccess
	db.ParticleUser
	UsedSize uint `json:"used_size"`
}

type ParticleStruct struct {
	APIBasicSuccess
	db.Particle
	Branches map[string][]ParticleBranchItem `json:"branches"`
}

type ParticleBranchItem struct {
	ID   uint   `json:"id"`
	Arch string `json:"arch"`
	Size uint   `json:"size"`
}

type RepatchGDServer struct {
	Name    string `json:"name" gorm:"column:srvName"`
	SrvId   string `json:"srvid" gorm:"column:srvid"`
	Players int    `json:"players" gorm:"column:userCount"`
	Levels  int    `json:"levels" gorm:"column:levelCount"`
	Icon    string `json:"icon" gorm:"column:icon"`
	Version string `json:"version" gorm:"column:version"`
	Recipe  string `json:"recipe" gorm:"column:recipe"`
}

type MinecraftCoresResponse struct {
	APIBasicSuccess
	Cores map[string]MCCore `json:"cores"`
	//AllVersions []string          `json:"all_versions"`
}

type APIServerMCCreateRequest struct {
	Name          string `json:"name"`
	SrvId         string `json:"srvid"`
	Tariff        string `json:"tariff"`
	Core          string `json:"core"`
	Version       string `json:"version"`
	AddStorage    int    `json:"add_storage"`
	DedicatedPort bool   `json:"dedicated_port"`
	Duration      string `json:"duration"`
	Promocode     string `json:"promocode"`
}
