package ent

import (
	"github.com/fruitspace/FiberAPI/providers"
	"github.com/fruitspace/FiberAPI/providers/ServerGD"
	"github.com/fruitspace/FiberAPI/providers/ServerMC"
	"github.com/fruitspace/FiberAPI/providers/particle"
	"github.com/fruitspace/FiberAPI/utils"
	"github.com/gofiber/fiber/v2"
)

type API struct {
	AccountProvider      *providers.AccountProvider
	NotificationProvider *providers.NotificationProvider
	PaymentProvider      *providers.PaymentProvider
	PromocodeProvider    *providers.PromocodeProvider
	ShopProvider         *providers.ShopProvider
	ServerGDProvider     *ServerGD.ServerGDProvider
	ServerMCProvider     *ServerMC.ServerMCProvider
	ParticleProvider     *particle.ParticleProvider

	SuperLock *utils.SuperLock
	Host      string
}

func (api *API) PerformAuth_(c *fiber.Ctx, acc *providers.Account) bool {
	token := c.Get("Authorization")
	if token == "" || !acc.GetUserBySession(token) {
		return false
	}
	if !acc.Data().IsActivated || acc.Data().IsBanned {
		return false
	}
	return true
}
