package db

import (
	gorm "github.com/cradio/gormx"
	"github.com/google/uuid"
	"time"
)

type MinecraftServer struct {
	UUID               uuid.UUID        `gorm:"primary_key;type:varchar(64);size:64;default:UUID()" json:"uuid"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	DeletedAt          gorm.DeletedAt   `gorm:"index" json:"deleted_at"`
	OwnerID            uint             `json:"owner_id"`
	MinecraftNetworkID string           `json:"minecraft_network_id"`
	MinecraftNetwork   MinecraftNetwork `json:"-"`
	Plan               string           `json:"plan"`
	ExpireDate         time.Time        `json:"expire_date"`
	AddDisk            int              `json:"add_disk"`
}
