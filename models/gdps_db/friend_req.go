package gdps_db

import (
	"time"
)

type FriendReq struct {
	ID         int       `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	UIDSrc     int       `gorm:"column:uid_src;type:int;" json:"uid_src"`
	UIDDest    int       `gorm:"column:uid_dest;type:int;" json:"uid_dest"`
	UploadDate time.Time `gorm:"column:uploadDate;type:datetime;" json:"upload_date"`
	Comment    string    `gorm:"column:comment;type:varchar;size:512;" json:"comment"`
	IsNew      int       `gorm:"column:isNew;type:tinyint;default:1;" json:"is_new"`
}

func (f *FriendReq) TableName() string {
	return "friendreqs"
}
