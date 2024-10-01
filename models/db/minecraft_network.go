package db

import (
	"encoding/json"
	gorm "github.com/cradio/gormx"
	"github.com/google/uuid"
	"time"
)

type MinecraftNetwork struct {
	UUID             uuid.UUID         `gorm:"primary_key;type:varchar(64);size:64;default:UUID()" json:"uuid"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	DeletedAt        gorm.DeletedAt    `gorm:"index" json:"deleted_at"`
	RegionID         uint              `json:"-"`
	Region           Region            `json:"-"`
	OwnerID          uint              `json:"owner_id"`
	MinecraftServers []MinecraftServer `json:"minecraft_servers"`
	Router           string            `json:"router"`
	RouterConfig     string            `gorm:"type:json;default:{}" json:"router_config"`
	Hostname         string            `json:"hostname"`
}

func (mn *MinecraftNetwork) GetRouterConfig() (conf MinecraftRouterConfig, err error) {
	err = json.Unmarshal([]byte(mn.RouterConfig), &conf)
	return
}

func (mn *MinecraftNetwork) SetRouterConfig(conf MinecraftRouterConfig) (err error) {
	cfg, err := json.Marshal(conf)
	if err == nil {
		mn.RouterConfig = string(cfg)
	}
	return
}

type MinecraftRouterConfig struct {
	Host           string   `json:"host"`
	Backends       []string `json:"backends"`
	OfflineMessage struct {
		Motd string `json:"motd"`
	} `json:"offline_message"`
	ProxyProtocol     bool `json:"proxy_protocol"`
	TCPShieldProtocol bool `json:"tcpshield_protocol"`
}
