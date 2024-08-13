package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/fruitspace/schemas/db/go/db"
	croc "github.com/m41denx/alligator"
	"io"
	"net/http"
	"strconv"
)

type PterodactylService struct {
	api       *croc.AppConfig
	wispToken string
}

func NewPterodactylService(apiKey string) *PterodactylService {
	app, _ := croc.NewApp("https://fruitspace.panel.gg", apiKey)
	return &PterodactylService{
		api: app,
	}

}

// region Users

func (p *PterodactylService) HasAccount(uid int) bool {
	_, err := p.api.GetUserByExternal(strconv.Itoa(uid))
	fmt.Println(err)
	return err == nil
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

// endregion

// region WISP

func (p *PterodactylService) WithWispToken(wispToken string) *PterodactylService {
	p.wispToken = wispToken
	return p
}

func (p *PterodactylService) GetWispNodes() (nodes []structs.WispNode, err error) {
	defer func() {
		if l := recover(); l != nil {
			utils.SendMessageDiscord(fmt.Sprintf("%+v", nodes))

		}
	}()
	url := "https://fruitspace.panel.gg/api/admin/nodes?page=1&per_page=25&sort=name&include[]=servers_count"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.wispToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var nodeResponse structs.WispNodesResponse
	err = json.Unmarshal(data, &nodeResponse)
	if err != nil {
		return
	}
	for _, obj := range nodeResponse.Data {
		nodes = append(nodes, obj.Attributes)
		fmt.Println(obj.Attributes)
	}
	return
}

func (p *PterodactylService) HasSuitableNodes(nodes []structs.WispNode, needRamGB uint, needCPU uint, needDiskGB uint) (ok bool) {
	for _, node := range nodes {
		freeCoresPercent := int(float64(node.Limits.CPU)*(1+float64(node.Limits.CPUOverallocate/100))) - node.CPUUsage
		freeRam := int(float64(node.Limits.Memory)*(1+float64(node.Limits.MemoryOverallocate/100))) - node.MemoryUsageMB
		freeDisk := int(float64(node.Limits.Disk)*(1+float64(node.Limits.DiskOverallocate/100))) - node.DiskUsageMB
		freeCores := uint(freeCoresPercent / 100)

		fmt.Printf("[%d] %d/%d cores, %d/%d ram, %d/%d disk\n", node.ID, freeCores, needCPU, freeRam, needRamGB*1024, freeDisk, needDiskGB*1024)

		if freeCores >= needCPU && uint(freeRam) >= needRamGB*1024 && uint(freeDisk) >= needDiskGB*1024 {
			return true
		}
	}
	return false
}

// endregion

// region Minecraft

func (p *PterodactylService) CreateMinecraftServer(name string, uid int, tariff structs.MCTariff, addDiskGB int, core structs.MCCore, version string) (srv croc.ServerAttributes, err error) {
	// Prepare egg env
	eggObj, err := p.api.GetEgg(5, core.EggID)
	if err != nil {
		return
	}
	egg := eggObj.Attributes
	if egg.DockerImage == "" {
		egg.DockerImage = "ghcr.io/pterodactyl/yolks:java_17"
	}
	env := make(map[string]string)
	for _, rel := range egg.Relationships.Variables.Data {
		d := rel.Attributes
		env[d.EnvVariable] = d.DefaultValue
	}
	env[core.VersionField] = version

	// Get Allocs
	allocID := 0
	nodes, err := p.api.GetNodes()
	if err != nil {
		return
	}
	for _, node := range nodes.Nodes {
		allocs, err := p.api.GetNodeAllocations(node.Attributes.ID)
		if err != nil {
			utils.SendMessageDiscord(fmt.Sprintf("%+v", err))
			continue
		}
		for _, alloc := range allocs.Allocations {
			if !alloc.Attributes.Assigned {
				allocID = alloc.Attributes.ID
				break
			}
		}
		if allocID != 0 {
			break
		}
	}

	if allocID == 0 {
		err = errors.New("no free allocations")
		return
	}

	// Get WISP UID
	user, err := p.api.GetUserByExternal(strconv.Itoa(uid))
	if err != nil {
		return
	}
	wispUid := user.Attributes.ID

	conf := croc.ServerChange{
		Name:        name,
		User:        wispUid,
		Egg:         core.EggID,
		DockerImage: egg.DockerImage,
		Startup:     tariff.GetStartupTemplate(),
		Environment: env,
		Limits: croc.ServerLimits{
			Memory: int(tariff.GetRAM() * 1024),
			Swap:   int(tariff.GetSwap() * 1024),
			Disk:   int(tariff.DiskGB*1024) + addDiskGB*1024,
			Io:     500,
			CPU:    int(tariff.CPU * 100),
		},
		FeatureLimits: croc.ServerFeatureLimits{
			Databases:   1,
			Allocations: 0,
		},
		Allocation: croc.ServerAllocation{
			Default:    allocID,
			Additional: nil,
		},
	}

	srvu, err := p.api.CreateServer(conf)
	if err != nil {
		return
	}
	srv = srvu.Attributes
	return
}

func (p *PterodactylService) GetMinecraftServerStatus(id int) (status bool, err error) {
	srv, err := p.api.GetServer(id)
	if err != nil {
		return
	}
	return !srv.Attributes.Suspended, nil
}

// endregion
