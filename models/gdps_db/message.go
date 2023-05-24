package gdps_db

import (
	"time"
)

type Message struct {
	ID         int       `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	UIDSrc     int       `gorm:"column:uid_src;type:int;" json:"uid_src"`
	UIDDest    int       `gorm:"column:uid_dest;type:int;" json:"uid_dest"`
	Subject    string    `gorm:"column:subject;type:varchar;size:256;" json:"subject"`
	Body       string    `gorm:"column:body;type:varchar;size:1024;" json:"body"`
	PostedTime time.Time `gorm:"column:postedTime;type:datetime;" json:"posted_time"`
	IsNew      int       `gorm:"column:isNew;type:tinyint;default:1;" json:"is_new"`
}

func (m *Message) TableName() string {
	return "messages"
}
