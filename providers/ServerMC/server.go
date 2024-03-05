package ServerMC

import (
	"errors"
	gorm "github.com/cradio/gormx"
	fiberapi "github.com/fruitspace/FiberAPI"
	"github.com/fruitspace/FiberAPI/models/db"
	"github.com/fruitspace/FiberAPI/models/structs"
	"github.com/fruitspace/FiberAPI/providers"
	"github.com/fruitspace/FiberAPI/services"
	"github.com/fruitspace/FiberAPI/utils"
	goversion "github.com/hashicorp/go-version"
	"strconv"
	"strings"
	"time"
)

type ServerMCProvider struct {
	db       *gorm.DB
	payments *providers.PaymentProvider
}

func NewServerMCProvider(db *gorm.DB, payments *providers.PaymentProvider) *ServerMCProvider {
	return &ServerMCProvider{
		db:       db,
		payments: payments,
	}
}

func (smp *ServerMCProvider) GetUserServers(uid int) (srvs []*db.ServerMc) {
	smp.db.Model(db.ServerMc{}).Where(db.ServerMc{OwnerID: uid}).Find(&srvs)
	return
}

func (smp *ServerMCProvider) CountServers() int {
	var cnt int64
	smp.db.Model(db.ServerMc{}).Count(&cnt)
	return int(cnt)
}

func (smp *ServerMCProvider) New() *ServerMC {
	return &ServerMC{
		p: smp,
	}
}

type ServerMC struct {
	p    *ServerMCProvider
	Data *db.ServerMc
}

func (s *ServerMC) CreateServer(user *db.User, name string, plan string, core string, version string, addStorage int, dedicPort bool, promocode string) (srvid string, err error) {
	pm := providers.NewPromocodeProvider(s.p.db)
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\n", "")
	tariff, ok := fiberapi.ProductMCTariffs[plan]
	if !ok {
		return "", errors.New("Invalid tariff |Tariff")
	}
	if addStorage > 20 || addStorage < 0 {
		return "", errors.New("Too much storage |inv_storage")
	}
	coreInfo, ok := structs.MCCoresEggs[core]
	if !ok {
		return "", errors.New("Invalid core |inv_core")
	}
	verConstraint, err := goversion.NewConstraint(coreInfo.VersionConstraint)
	if err != nil {
		return "gover_newconstraint", errors.New("Invalid version |inv_version")
	}

	if ver, err := goversion.NewVersion(version); err != nil || !verConstraint.Check(ver) {
		return "gover_check", errors.New("Invalid version |inv_version")
	}

	when := time.Now().AddDate(0, 1, 0)

	price := tariff.PriceRUB + addStorage*fiberapi.ProductMCAdditionalDiskPricePer10GB

	if promocode != "" {
		promo := pm.Get(promocode)
		if promo == nil {
			return "", errors.New("Invalid promocode |promo_invalid")
		}
		prc, err := promo.Use(float64(price), "mc", plan)
		if err != nil {
			return "promo_use", err
		}
		price = int(prc)
	}

	addStorage *= 10

	cs := db.ServerMc{
		SrvName:    name,
		Plan:       plan,
		OwnerID:    user.UID,
		ExpireDate: when,
		Version:    version,
		Core:       core,
		RamMin:     int(tariff.MinRamGB),
		RamMax:     int(tariff.MaxRamGB),
		CPUs:       int(tariff.CPU),
		AddDisk:    addStorage,
	}

	ptero := services.NewPterodactylService(utils.GetEnv("PTERO_TOKEN", "")).WithWispToken(utils.GetEnv("WISP_TOKEN", ""))

	// Check Server Availability
	nodes, err := ptero.GetWispNodes()
	if err != nil {
		return "ptero_getnodes", errors.New("No nodes available |no_nodes")
	}
	if !ptero.HasSuitableNodes(nodes, tariff.GetRAM(), tariff.CPU, uint(addStorage)+tariff.DiskGB+tariff.GetSwap()) {
		return "ptero_findnodes", errors.New("No nodes available |no_nodes")
	}

	resp := s.p.payments.SpendMoney(user.UID, float64(price))
	if resp.Status != "ok" {
		return "", errors.New(resp.Message)
	}

	if !ptero.HasAccount(user.UID) {
		err = ptero.CreateAccount(*user, "") // TODO: Check if password is sent to email
		if err != nil {
			return "ptero_createacc", errors.New("Cannot create account |wisp_account")
		}
	}

	server, err := ptero.CreateMinecraftServer(name, user.UID, tariff, addStorage, coreInfo, version)
	if err != nil {
		return "ptero_createserv", errors.New("Cannot create server |wisp_server")
	}
	cs.SrvID = strconv.Itoa(server.ID)
	err = s.p.db.Model(db.ServerMc{}).Create(&cs).Error
	if err != nil {
		return "mcdb", errors.New("Cannot create server |wisp_server")
	}
	return cs.SrvID, nil
}
