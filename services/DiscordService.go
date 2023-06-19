package services

import (
	"bytes"
	"encoding/json"
	"github.com/fruitspace/FiberAPI/utils"
	"io"
	"net/http"
)

type DiscordService struct {
	config map[string]string
}

func NewDiscordService(conf map[string]string) *DiscordService {
	return &DiscordService{config: conf}
}

func (d *DiscordService) AuthByCode(code string) (*DiscordAuthResponse, error) {
	params, _ := json.Marshal(map[string]string{
		"client_id":     d.config["appid"],
		"client_secret": d.config["secret"],
		"redirect_uri":  d.config["url"],
		"grant_type":    "authorization_code",
		"code":          code,
	})
	var data map[string]interface{}

	resp, err := http.Post("https://discord.com/api/oauth2/token", "application/json", bytes.NewBuffer(params))
	if utils.Should(err) != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
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
