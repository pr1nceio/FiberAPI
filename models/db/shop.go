package db

import (
	"database/sql"
)

type Shop struct {
	ID          int            `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	Name        string         `gorm:"column:name;type:varchar;size:255;" json:"name"`
	OwnerID     int            `gorm:"column:owner_id;type:int;" json:"owner_id"`
	Balance     float64        `gorm:"column:balance;type:double;default:0.00;" json:"balance"`
	BalanceHash string         `gorm:"column:balanceHash;type:varchar;size:255;" json:"-"`
	Items       sql.NullString `gorm:"column:items;type:json;" json:"items"`
}

func (s *Shop) TableName() string {
	return "shops"
}
