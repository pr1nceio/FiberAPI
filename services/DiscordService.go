package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fruitspace/FiberAPI/utils"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type DiscordService struct {
	config map[string]string
}

func NewDiscordService(conf map[string]string) *DiscordService {
	return &DiscordService{config: conf}
}

func (d *DiscordService) AuthByCode(code string) (*DiscordAuthResponse, error) {
	lo := url.Values{}
	lo.Add("client_id", d.config["appid"])
	lo.Add("client_secret", d.config["secret"])
	lo.Add("redirect_uri", d.config["url"])
	lo.Add("grant_type", "authorization_code")
	lo.Add("code", code)
	var data map[string]interface{}

	resp, err := http.Post("https://discord.com/api/oauth2/token", "application/x-www-form-urlencoded", strings.NewReader(lo.Encode()))
	if utils.Should(err) != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &data)
	log.Println(string(body))
	if err != nil {
		return nil, err
	}
	if serr, ok := data["error"].(string); ok {
		return nil, errors.New(fmt.Sprintf("%s: %s", serr, data["error_description"].(string)))
	}
	mt := DiscordAuthResponse{
		Token:        data["access_token"].(string),
		RefreshToken: data["refresh_token"].(string),
	}
	authReq, _ := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	authReq.Header.Set("Authorization", "Bearer "+mt.Token)
	authResp, err := http.DefaultClient.Do(authReq)
	if err != nil {
		return nil, err
	}
	var authData map[string]interface{}
	authBody, err := io.ReadAll(authResp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(authBody, &authData)
	if err != nil {
		return nil, err
	}
	mt.ClientID = authData["id"].(string)

	return &mt, nil
}

type DiscordAuthResponse struct {
	ClientID     string
	Token        string
	RefreshToken string
}
