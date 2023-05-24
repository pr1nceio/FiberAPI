package gdps_db

import (
	"time"
)

type Comment struct {
	ID         int       `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	UID        int       `gorm:"column:uid;type:int;" json:"uid"`
	LvlID      int       `gorm:"column:lvl_id;type:int;" json:"lvl_id"`
	Comment    string    `gorm:"column:comment;type:varchar;size:128;" json:"comment"`
	PostedTime time.Time `gorm:"column:postedTime;type:datetime;" json:"posted_time"`
	Likes      int       `gorm:"column:likes;type:int;default:0;" json:"likes"`
	IsSpam     int       `gorm:"column:isSpam;type:tinyint;default:0;" json:"is_spam"`
	Percent    int       `gorm:"column:percent;type:tinyint;" json:"percent"`
}

func (c *Comment) TableName() string {
	return "comments"
}
