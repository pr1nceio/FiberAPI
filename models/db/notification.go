package db

import (
	"time"
)

type Notification struct {
	UUID       string    `gorm:"primary_key;column:uuid;type:varchar;size:64;" json:"uuid"`
	TargetUID  int       `gorm:"column:target_uid;type:int;" json:"target_uid"`
	Title      string    `gorm:"column:title;type:varchar;size:64;" json:"title"`
	Text       string    `gorm:"column:text;type:varchar;size:128;" json:"text"`
	CreatedAt  time.Time `gorm:"column:sendDate;type:datetime;default:CURRENT_TIMESTAMP;" json:"send_date"`
	ExpireDate time.Time `gorm:"column:expireDate;type:datetime;default:CURRENT_TIMESTAMP;" json:"expire_date"`
	UserRead   bool      `gorm:"column:userRead;type:tinyint;default:0;" json:"user_read"`
}

func (n *Notification) TableName() string {
	return "notifications"
}
