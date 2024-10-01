package ServerMC

import (
	gorm "github.com/cradio/gormx"
	"github.com/fruitspace/HyprrSpace/models/db"
	"github.com/fruitspace/HyprrSpace/providers"
	"github.com/m41denx/alligator"
)

type ServerMCProvider struct {
	db       *gorm.DB
	payments *providers.PaymentProvider
	pteroapi *alligator.Application
}

func NewServerMCProvider(db *gorm.DB, payments *providers.PaymentProvider) *ServerMCProvider {
	return &ServerMCProvider{
		db:       db,
		payments: payments,
	}
}

func (smp *ServerMCProvider) WithPterodactylAPI(token string) *ServerMCProvider {
	smp.pteroapi, _ = alligator.NewApp("https://panel.fruitspace.one", token)
	return smp
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

func (smp *ServerMCProvider) ListRegions() (region []db.RegionPublic) {
	smp.db.Model(db.Region{}).Find(&region)
	return
}

func (smp *ServerMCProvider) GetPricing(regionID int) (prices []db.PricingPublic) {
	smp.db.Model(db.Pricing{}).Preload("Tariff").Where(db.Pricing{RegionID: uint(regionID)}).Find(&prices)
	return
}

func (smp *ServerMCProvider) New() *ServerMC {
	return &ServerMC{
		p: smp,
	}
}
