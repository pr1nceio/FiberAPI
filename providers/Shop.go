package providers

import (
	"fmt"
	"github.com/cradio/gormx"
	"github.com/fruitspace/schemas/db/go/db"
)

//region ShopProvider

type ShopProvider struct {
	db *gorm.DB
}

func NewShopProvider(db *gorm.DB) *ShopProvider {
	return &ShopProvider{db: db}
}

func (sp *ShopProvider) GetUserShopsBalance(uid int) (sum float64) {
	sp.db.Where(db.Shop{OwnerID: uid}).Select(
		fmt.Sprintf("SUM(%s)", gorm.Column(db.Shop{}, "Balance")),
	).First(&sum)
	return sum
}

func (sp *ShopProvider) New() *Shop {
	return &Shop{shop: &db.Shop{}, p: sp}
}

//endregion

//region Shop

type Shop struct {
	shop *db.Shop
	p    *ShopProvider
}

func (s *Shop) GetShopById(id int) bool {
	return s.p.db.First(&s.shop, id).Error == nil
}

//endregion
