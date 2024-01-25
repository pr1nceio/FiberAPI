package services

import (
	"encoding/json"
	"fmt"
	"github.com/fruitspace/FiberAPI/models/db"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/utils"
	croc "github.com/parkervcp/crocgodyl"
	"io"
	"net/http"
	"strconv"
)

type PterodactylService struct {
	api       *croc.AppConfig
	wispToken string
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

// region Users

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

func (p *PterodactylService) ExistsAccount(uid int) bool {
	_, err := p.api.GetUserByExternal(strconv.Itoa(uid))
	return err == nil
}

// endregion

// region WISP

func (p *PterodactylService) WithWispToken(wispToken string) *PterodactylService {
	p.wispToken = wispToken
	return p
}

func (p *PterodactylService) GetWispNodes() (nodes []*structs.WispNode, err error) {
	defer func() {
		if l := recover(); l != nil {
			utils.SendMessageDiscord(fmt.Sprintf("%+v", nodes))

		}
	}()
	url := "https://fruitspace.panel.gg/api/admin/nodes?page=1&per_page=25&sort=name&include[]=servers_count"
	req, _ := http.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{
		Name:  "wisp_panel_session",
		Value: p.wispToken,
	})

	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var nodeResponse structs.WispNodesResponse
	err = json.Unmarshal(data, &nodeResponse)
	if err != nil {
		return
	}
	for _, obj := range nodeResponse.Data {
		nodes = append(nodes, obj.AsWispNode())
	}
	return
}

func (p *PterodactylService) HasSuitableNodes(nodes []*structs.WispNode, needRam uint, needCPU uint, needDisk uint) (ok bool) {
	for _, node := range nodes {
		freeCoresPercent := int(float64(node.Limits.CPU)*(1+float64(node.Limits.CPUOverallocate/100))) - node.CPUUsage
		freeRam := int(float64(node.Limits.Memory)*(1+float64(node.Limits.MemoryOverallocate/100))) - node.MemoryUsageMB
		freeDisk := int(float64(node.Limits.Disk)*(1+float64(node.Limits.DiskOverallocate/100))) - node.DiskUsageMB
		freeCores := uint(freeCoresPercent / 100)

		if freeCores >= needCPU && uint(freeRam) >= needRam && uint(freeDisk) >= needDisk {
			return true
		}
	}
	return false
}

// endregion

// region Minecraft

func (p *PterodactylService) CreateMinecraftServer() {
}

// endregion
