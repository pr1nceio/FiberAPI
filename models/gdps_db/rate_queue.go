package gdps_db

type RateQueue struct {
	ID     int32  `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	LvlID  int32  `gorm:"column:lvl_id;type:int;" json:"lvl_id"`
	Name   string `gorm:"column:name;type:varchar;size:32;" json:"name"`
	UID    int32  `gorm:"column:uid;type:int;" json:"uid"`
	ModUID int32  `gorm:"column:mod_uid;type:int;" json:"mod_uid"`
}

func (r *RateQueue) TableName() string {
	return "rateQueue"
}
