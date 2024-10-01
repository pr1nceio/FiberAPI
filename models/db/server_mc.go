package db

import "time"

type ServerMc struct {
	ID          int       `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"-"`
	SrvID       string    `gorm:"column:srvid;type:varchar(8);size:8;" json:"srvid"`
	SrvName     string    `gorm:"column:srv_name;type:varchar(255);size:255;" json:"srv_name"`
	Plan        string    `gorm:"column:plan;type:varchar(8);size:8;" json:"plan"`
	OwnerID     int       `gorm:"column:owner_id;type:int;" json:"owner_id"`
	CreatedAt   time.Time `gorm:"column:creation_date;type:datetime;default:CURRENT_TIMESTAMP;" json:"creation_date"`
	ExpireDate  time.Time `gorm:"column:expire_date;type:datetime;default:CURRENT_TIMESTAMP;" json:"expire_date"`
	Version     string    `gorm:"column:version;type:varchar(32);size:32;" json:"version"`
	Core        string    `gorm:"column:core;type:varchar(64);size:64;" json:"core"`
	RamMin      int       `gorm:"column:ram_min;type:int;" json:"ram_min"`
	RamMax      int       `gorm:"column:ram_max;type:int;" json:"ram_max"`
	CPUs        int       `gorm:"column:cpus;type:int;" json:"cpus"`
	AddDisk     int       `gorm:"column:add_disk;type:int;" json:"add_disk"`
	Description string    `gorm:"column:description;type:varchar(255);size:255;" json:"description"`
	Icon        string    `gorm:"column:icon;type:varchar(16);size:16;default:mc_default.png;" json:"icon"`
}

func (ServerMc) TableName() string {
	return "servers_mc"
}
