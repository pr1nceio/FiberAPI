package db

import gorm "github.com/cradio/gormx"

type Pricing struct {
	gorm.Model
	RegionID uint
	Region   Region
	TariffID uint
	Tariff   Tariff
	Price    uint
}

type PricingPublic struct {
	ID       uint   `json:"id"`
	RegionID uint   `json:"-"`
	Region   Region `json:"-"`
	TariffID uint   `json:"-"`
	Tariff   Tariff `json:"tariff"`
	Price    uint   `json:"price"`
}
