package db

type ServerMc struct {
	ID          int    `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"-"`
	SrvID       string `gorm:"column:srvid;type:varchar;size:8;" json:"srvid"`
	SrvName     string `gorm:"column:srvName;type:varchar;size:32;" json:"srv_name"`
	Plan        string `gorm:"column:plan;type:varchar;size:8;" json:"plan"`
	OwnerID     int    `gorm:"column:owner_id;type:int;" json:"owner_id"`
	Version     string `gorm:"column:version;type:varchar;size:16;" json:"version"`
	Core        string `gorm:"column:core;type:varchar;size:64;" json:"core"`
	RamMin      int    `gorm:"column:ramMin;type:int;" json:"ram_min"`
	RamMax      int    `gorm:"column:ramMax;type:int;" json:"ram_max"`
	CPUs        int    `gorm:"column:cpus;type:int;" json:"cpus"`
	SSD         int    `gorm:"column:ssd;type:int;" json:"ssd"`
	Description string `gorm:"column:description;type:varchar;size:1000;" json:"description"`
	Icon        string `gorm:"column:icon;type:varchar;size:16;default:mc_default.png;" json:"icon"`
}
