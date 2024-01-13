package services

import (
	"github.com/fruitspace/FiberAPI/models/db"
	croc "github.com/parkervcp/crocgodyl"
	"strconv"
)

type PterodactylService struct {
	api *croc.AppConfig
}

func NewPterodactylService(apiKey string) (*PterodactylService, error) {
	app, err := croc.NewApp("https://fruitspace.panel.gg", apiKey)
	if err != nil {
		return nil, err
	}
	return &PterodactylService{
		api: app,
	}, nil

}

func (p *PterodactylService) CreateAccount(user db.User, password string) (err error) {
	_, err = p.api.CreateUser(croc.UserAttributes{
		Username:   user.Uname,
		Email:      user.Email,
		FirstName:  user.Name,
		LastName:   user.Surname,
		Password:   password,
		ExternalID: strconv.Itoa(user.UID),
	})
	return err
}
