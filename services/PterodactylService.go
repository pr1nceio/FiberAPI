package services

import (
	"errors"
	"fmt"
	"github.com/fruitspace/HyprrSpace/models/db"
	"github.com/fruitspace/HyprrSpace/models/structs"
	"github.com/fruitspace/HyprrSpace/utils"
	gator "github.com/m41denx/alligator"
	"github.com/m41denx/alligator/options"
	"strconv"
)

type PterodactylService struct {
	api *gator.Application
}

// ptla_ItoKeY0gvzIfhaLxcJxNNogOTBJKUGshAclNpYQynSF
func NewPterodactylService(apiKey string) *PterodactylService {
	app, _ := gator.NewApp("https://panel.fruitspace.one", apiKey)
	return &PterodactylService{
		api: app,
	}

}

// region Users

func (p *PterodactylService) HasAccount(uid int) bool {
	_, err := p.api.GetUserExternal(strconv.Itoa(uid))
	fmt.Println(err)
	return err == nil
}

func (p *PterodactylService) CreateAccount(user db.User, password string) (err error) {
	_, err = p.api.CreateUser(gator.CreateUserDescriptor{
		Username:   user.Uname,
		Email:      user.Email,
		FirstName:  user.Name,
		LastName:   "@FruitSpace",
		Password:   password,
		ExternalID: strconv.Itoa(user.UID),
	})
	return err
}

// endregion

// region Nodes

func (p *PterodactylService) GetNodes() (nodes []structs.PterodactylNodeExtended, err error) {
	defer func() {
		if l := recover(); l != nil {
			utils.SendMessageDiscord(fmt.Sprintf("%#v", nodes))

		}
	}()

	nodeResp, err := p.api.ListNodes(options.ListNodesOptions{Include: options.IncludeNodes{Servers: true}})
	if err != nil {
		return
	}
	for _, obj := range nodeResp {
		nodeAttr := structs.PterodactylNodeExtended{
			Node:         obj,
			ServersCount: int64(len(obj.Servers)),
		}
		for _, srv := range obj.Servers {
			nodeAttr.CPUUsage += srv.Limits.CPU
			nodeAttr.MemoryUsageMB += srv.Limits.Memory
			nodeAttr.DiskUsageMB += srv.Limits.Disk
		}
		nodes = append(nodes, nodeAttr)
	}
	return
}

// calculate percentage of total free resources (avg of cpu, ram and disk)

func (p *PterodactylService) HasSuitableNodes(nodes []structs.PterodactylNodeExtended, needRamGB uint, needCPU uint, needDiskGB uint) (ok bool) {
	for _, node := range nodes {
		//TODO: how the fuck are we going to fetch free cores count ohmygoooood
		freeCoresPercent := 2400 //int(float64(node.Limits.CPU)*(1+float64(node.Limits.CPUOverallocate/100))) - node.CPUUsage
		freeRam := int64(float64(node.Memory)*(1+float64(node.MemoryOverallocate/100))) - node.MemoryUsageMB
		freeDisk := int64(float64(node.Disk)*(1+float64(node.DiskOverallocate/100))) - node.DiskUsageMB
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

func (p *PterodactylService) CreateMinecraftServer(name string, uid int, tariff structs.MCTariff, addDiskGB int, core structs.MCCore, version string) (srv *gator.AppServer, err error) {
	// Prepare egg env
	egg, err := p.api.GetEgg(1, core.EggID)
	if err != nil {
		return
	}
	if egg.DockerImage == "" {
		egg.DockerImage = "ghcr.io/pterodactyl/yolks:java_17"
	}
	env := make(map[string]string)
	for _, rel := range egg.Variables {
		env[rel.EnvVariable] = rel.DefaultValue
	}
	env[core.VersionField] = version

	// Get Allocs
	allocID := 0
	nodes, err := p.api.ListNodes(options.ListNodesOptions{
		Include: options.IncludeNodes{Allocations: true},
	})
	if err != nil {
		return
	}
	for _, node := range nodes {
		for _, alloc := range node.Allocations {
			if !alloc.Assigned {
				allocID = alloc.ID
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

	// Get Pterodactyl UID
	user, err := p.api.GetUserExternal(strconv.Itoa(uid))
	if err != nil {
		return
	}

	envx := make(map[string]interface{})
	for k, v := range env {
		envx[k] = v
	}

	conf := gator.CreateServerDescriptor{
		Name:        name,
		User:        user.ID,
		Egg:         core.EggID,
		DockerImage: egg.DockerImage,
		Startup:     tariff.GetStartupTemplate(),
		Environment: envx,
		Limits: &gator.Limits{
			Memory: int64(tariff.GetRAM() * 1024),
			Swap:   int64(tariff.GetSwap() * 1024),
			Disk:   int64(tariff.DiskGB*1024) + int64(addDiskGB*1024),
			IO:     500,
			CPU:    int64(tariff.CPU * 100),
		},
		FeatureLimits: gator.FeatureLimits{
			Databases:   1,
			Allocations: 0,
		},
		Allocation: &gator.AllocationDescriptor{
			Default:    allocID,
			Additional: nil,
		},
	}

	return p.api.CreateServer(conf)
}

func (p *PterodactylService) GetMinecraftServerStatus(id int) (status bool, err error) {
	srv, err := p.api.GetServer(id)
	if err != nil {
		return
	}
	return !srv.Suspended, nil
}

// endregion
