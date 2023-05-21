package gdps_db

type Friendship struct {
	ID    int32 `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	UID1  int32 `gorm:"column:uid1;type:int;" json:"uid_1"`
	UID2  int32 `gorm:"column:uid2;type:int;" json:"uid_2"`
	U1New int32 `gorm:"column:u1_new;type:tinyint;default:1;" json:"u_1_new"`
	U2New int32 `gorm:"column:u2_new;type:tinyint;default:1;" json:"u_2_new"`
}

func (f *Friendship) TableName() string {
	return "friendships"
}
