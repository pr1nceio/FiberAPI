package gdps_db

import (
	"time"
)

type Quest struct {
	ID         int32     `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	Type       int32     `gorm:"column:type;type:tinyint;" json:"type"`
	Name       string    `gorm:"column:name;type:varchar;size:64;" json:"name"`
	Needed     int32     `gorm:"column:needed;type:int;default:0;" json:"needed"`
	Reward     int32     `gorm:"column:reward;type:int;default:0;" json:"reward"`
	LvlID      int32     `gorm:"column:lvl_id;type:int;default:0;" json:"lvl_id"`
	TimeExpire time.Time `gorm:"column:timeExpire;type:datetime;" json:"time_expire"`
}

func (q *Quest) TableName() string {
	return "quests"
}
