package gdps_db

import (
	"time"
)

type Action struct {
	ID       int32     `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	Date     time.Time `gorm:"column:date;type:datetime;" json:"date"`
	UID      int32     `gorm:"column:uid;type:int;" json:"uid"`
	Type     int32     `gorm:"column:type;type:tinyint;" json:"type"`
	TargetID int32     `gorm:"column:target_id;type:int;" json:"target_id"`
	IsMod    int32     `gorm:"column:isMod;type:tinyint;default:0;" json:"is_mod"`
	Data     string    `gorm:"column:data;type:text;size:4294967295;default:{};" json:"data"`
}

func (a *Action) TableName() string {
	return "actions"
}
