package providers

import (
	"errors"
	"github.com/fruitspace/FiberAPI/models/db"
	"gorm.io/gorm"
	"math"
	"strings"
	"time"
)

//region PromocodeProvider

type PromocodeProvider struct {
	db *gorm.DB
}

func NewPromocodeProvider(db *gorm.DB) *PromocodeProvider {
	return &PromocodeProvider{db: db}
}

func (pp *PromocodeProvider) Get(code string) *Promocode {
	var p db.Promocode
	pp.db.Where(db.Promocode{Code: code}).First(&p)
	return &Promocode{p: pp, promocode: &p}
}

//endregion

//region Promocode

type Promocode struct {
	p         *PromocodeProvider
	promocode *db.Promocode
}

func (p *Promocode) GetType() string {
	return strings.Split(p.promocode.ProdType, ":")[0]
}

func (p *Promocode) GetPlan() string {
	return strings.Split(p.promocode.ProdType, ":")[1]
}

func (p *Promocode) CheckPlan(uPlan string) bool {
	plan := p.GetPlan()
	return plan == "-1" || plan == uPlan
}

func (p *Promocode) ApplyDiscount(price float64) float64 {
	p.promocode.Uses++
	p.p.db.Model(&p.promocode).Updates(db.Promocode{Uses: p.promocode.Uses})
	return math.Round(price*float64(100-p.promocode.Discount)) / 100
}

// Use returns new price and error if any
func (p *Promocode) Use(price float64, product string, plan string) (float64, error) {

	if p.promocode.ExpireDate.Before(time.Now()) {
		return 0, errors.New("Promocode expired |promo_expire")
	}
	if p.promocode.Uses >= p.promocode.MaxUses && p.promocode.MaxUses != -1 {
		return 0, errors.New("Promocode uses limit reached |promo_limit")
	}
	if p.promocode.ProdType != "" && (p.GetType() != product || !p.CheckPlan(plan)) {
		return 0, errors.New("Invalid promocode |promo_invalid")
	}
	return p.ApplyDiscount(price), nil
}

//endregion
