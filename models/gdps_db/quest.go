package gdps_db

import (
	"time"
)

type Quest struct {
	ID         int       `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	Type       int       `gorm:"column:type;type:tinyint;" json:"type"`
	Name       string    `gorm:"column:name;type:varchar;size:64;" json:"name"`
	Needed     int       `gorm:"column:needed;type:int;default:0;" json:"needed"`
	Reward     int       `gorm:"column:reward;type:int;default:0;" json:"reward"`
	LvlID      int       `gorm:"column:lvl_id;type:int;default:0;" json:"lvl_id"`
	TimeExpire time.Time `gorm:"column:timeExpire;type:datetime;" json:"time_expire"`
}

func (q *Quest) TableName() string {
	return "quests"
}
