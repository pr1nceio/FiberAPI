package gdps_db

import (
	"github.com/fruitspace/FiberAPI/models"
	"time"
)

type Action struct {
	ID       int            `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	Date     time.Time      `gorm:"column:date;type:datetime;" json:"date"`
	UID      int            `gorm:"column:uid;type:int;" json:"uid"`
	Type     int            `gorm:"column:type;type:tinyint;" json:"type"`
	TargetID int            `gorm:"column:target_id;type:int;" json:"target_id"`
	IsMod    int            `gorm:"column:isMod;type:tinyint;default:0;" json:"is_mod"`
	Data     models.JSONMap `gorm:"column:data;type:json;size:4294967295;default:{};" json:"data"`
}

func (a *Action) TableName() string {
	return "actions"
}
