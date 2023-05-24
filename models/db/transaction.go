package db

import (
	"time"
)

type Transaction struct {
	ID        int       `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	PayID     string    `gorm:"column:pay_id;type:varchar;size:32;default:0;" json:"pay_id"`
	UID       int       `gorm:"column:uid;type:int;" json:"uid"`
	Amount    float64   `gorm:"column:amount;type:double;" json:"amount"`
	CreatedAt time.Time `gorm:"column:date;type:datetime;default:CURRENT_TIMESTAMP;" json:"date"`
	IsActive  bool      `gorm:"column:isActive;type:tinyint;default:1;" json:"is_active"`
	Method    string    `gorm:"column:method;type:varchar;size:32;default:xx;" json:"method"`
	GoPayURL  string    `gorm:"column:goPayUrl;type:varchar;size:1024;" json:"go_pay_url"`
}

func (t *Transaction) TableName() string {
	return "transactions"
}
