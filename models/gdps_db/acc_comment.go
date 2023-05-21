package gdps_db

import (
	"time"
)

type AccComment struct {
	ID         int32     `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	UID        int32     `gorm:"column:uid;type:int;" json:"uid"`
	Comment    string    `gorm:"column:comment;type:varchar;size:128;" json:"comment"`
	PostedTime time.Time `gorm:"column:postedTime;type:datetime;" json:"posted_time"`
	Likes      int32     `gorm:"column:likes;type:int;default:0;" json:"likes"`
	IsSpam     int32     `gorm:"column:isSpam;type:tinyint;default:0;" json:"is_spam"`
}

func (a *AccComment) TableName() string {
	return "acccomments"
}
