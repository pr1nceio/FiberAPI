package db

type MachineTelemetry struct {
	GUID  string `gorm:"primary_key;column:guid;type:varchar(64);size:64;" json:"guid"`
	Os    string `gorm:"column:os;type:varchar(512);size:512;default:Generic Windows OS;" json:"os"`
	CPU   string `gorm:"column:cpu;type:varchar(128);size:128;default:Unknown;" json:"cpu"`
	Cores int    `gorm:"column:cores;type:int;default:0;" json:"cores"`
	RAM   int    `gorm:"column:ram;type:int;default:0;" json:"ram"`
	Gpus  string `gorm:"column:gpus;type:text;size:65535;" json:"gpus"`
	Av    string `gorm:"column:av;type:varchar(1024);size:1024;default:UNKNOWN;" json:"av"`
}

func (m *MachineTelemetry) TableName() string {
	return "MachineTelemetry"
}
