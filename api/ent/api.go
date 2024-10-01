package ent

import (
	"github.com/fruitspace/HyprrSpace/providers"
	"github.com/fruitspace/HyprrSpace/providers/ServerGD"
	"github.com/fruitspace/HyprrSpace/providers/ServerMC"
	"github.com/fruitspace/HyprrSpace/providers/particle"
	"github.com/fruitspace/HyprrSpace/utils"
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
	token := c.Cookies("token")
	if len(token) == 0 {
		token = c.Get("Authorization")
	}
	if token == "" || !acc.GetUserBySession(token) {
		return false
	}
	if !acc.Data().IsActivated || acc.Data().IsBanned {
		return false
	}
	return true
}

func (api *API) SetToken_(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		Domain:   "fruitspace.one",
		MaxAge:   1000 * 60 * 60 * 24 * 30,
		Secure:   true,
		HTTPOnly: true,
	})
}
