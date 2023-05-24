package gdps_db

import (
	"time"
)

type Score struct {
	ID         int       `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	UID        int       `gorm:"column:uid;type:int;" json:"uid"`
	LvlID      int       `gorm:"column:lvl_id;type:int;" json:"lvl_id"`
	PostedTime time.Time `gorm:"column:postedTime;type:datetime;" json:"posted_time"`
	Percent    int       `gorm:"column:percent;type:tinyint;" json:"percent"`
	Attempts   int       `gorm:"column:attempts;type:int;default:0;" json:"attempts"`
	Coins      int       `gorm:"column:coins;type:tinyint;default:0;" json:"coins"`
}

func (s *Score) TableName() string {
	return "scores"
}
