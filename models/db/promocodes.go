package db

import (
	"time"
)

type Promocode struct {
	ID         int       `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	Code       string    `gorm:"column:code;type:varchar(64);size:64;" json:"code"`
	Discount   int       `gorm:"column:discount;type:int;" json:"discount"`
	ProdType   string    `gorm:"column:prodType;type:varchar(8);size:8;" json:"prod_type"`
	ExpireDate time.Time `gorm:"column:expireDate;type:datetime;default:CURRENT_TIMESTAMP;" json:"expire_date"`
	Uses       int       `gorm:"column:uses;type:int;default:0;" json:"uses"`
	MaxUses    int       `gorm:"column:maxUses;type:int;default:1;" json:"max_uses"`
}

func (p *Promocode) TableName() string {
	return "promocodes"
}
