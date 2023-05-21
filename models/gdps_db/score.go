package gdps_db

import (
	"time"
)

type Score struct {
	ID         int32     `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	UID        int32     `gorm:"column:uid;type:int;" json:"uid"`
	LvlID      int32     `gorm:"column:lvl_id;type:int;" json:"lvl_id"`
	PostedTime time.Time `gorm:"column:postedTime;type:datetime;" json:"posted_time"`
	Percent    int32     `gorm:"column:percent;type:tinyint;" json:"percent"`
	Attempts   int32     `gorm:"column:attempts;type:int;default:0;" json:"attempts"`
	Coins      int32     `gorm:"column:coins;type:tinyint;default:0;" json:"coins"`
}

func (s *Score) TableName() string {
	return "scores"
}
