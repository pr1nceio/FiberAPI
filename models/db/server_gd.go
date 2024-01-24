package db

import (
	"database/sql"
	"github.com/fruitspace/FiberAPI/models"
	"time"
)

type ServerGd struct {
	ID               int            `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"-"`
	SrvID            string         `gorm:"column:srvid;type:varchar;size:8;" json:"srvid"`
	SrvName          string         `gorm:"column:srvName;type:varchar;size:32;" json:"srv_name"`
	Plan             int            `gorm:"column:plan;type:tinyint;" json:"plan"`
	OwnerID          int            `gorm:"column:owner_id;type:int;" json:"owner_id"`
	DbPassword       string         `gorm:"column:dbPassword;type:varchar;size:64;" json:"db_password"`
	SrvKey           string         `gorm:"column:srvKey;type:varchar;size:32;" json:"srv_key"`
	UserCount        int            `gorm:"column:userCount;type:int;default:0;" json:"user_count"`
	LevelCount       int            `gorm:"column:levelCount;type:int;default:0;" json:"level_count"`
	CommentCount     int            `gorm:"column:commentCount;type:int;default:0;" json:"comment_count"`
	PostCount        int            `gorm:"column:postCount;type:int;default:0;" json:"post_count"`
	CreatedAt        time.Time      `gorm:"column:creationDate;type:datetime;default:CURRENT_TIMESTAMP;" json:"creation_date"`
	ExpireDate       time.Time      `gorm:"column:expireDate;type:datetime;default:CURRENT_TIMESTAMP;" json:"expire_date"`
	ClientAndroidURL string         `gorm:"column:clientAndroidURL;type:varchar;size:255;" json:"client_android_url"`
	ClientIOSURL     string         `gorm:"column:clientIOSURL;type:varchar;size:255;" json:"client_ios_url"`
	ClientWindowsURL string         `gorm:"column:clientWindowsURL;type:varchar;size:255;" json:"client_windows_url"`
	ClientMacOSURL   string         `gorm:"column:clientMacOSURL;type:varchar;size:255;" json:"client_macos_url"`
	AutoPay          int            `gorm:"column:autoPay;type:tinyint;default:0;" json:"auto_pay"`
	Backups          sql.NullString `gorm:"column:backups;type:json;" json:"backups"`
	MStatHistory     models.JSONMap `gorm:"column:mStatHistory;type:json;" json:"m_stat_history"`
	Icon             string         `gorm:"column:icon;type:varchar;size:16;default:gd_default.png;" json:"icon"`
	Description      string         `gorm:"column:description;type:varchar;size:1000;default:Welcome to my GDPS!;" json:"description"`
	TextAlign        int            `gorm:"column:textAlign;type:tinyint;default:0;" json:"text_align"`
	Visits           int            `gorm:"column:visits;type:int;default:0;" json:"visits"`
	Discord          string         `gorm:"column:discord;type:varchar;size:128;" json:"discord"`
	Vk               string         `gorm:"column:vk;type:varchar;size:128;" json:"vk"`
	IsSpaceMusic     bool           `gorm:"column:isSpaceMusic;type:tinyint;default:0;" json:"is_space_music"`
	Is22             bool           `gorm:"column:is22;type:tinyint;default:0;" json:"is_22"`
	IsCustomTextures bool           `gorm:"column:isCustomTextures;type:tinyint;default:0;" json:"is_custom_textures"`
	Version          string         `gorm:"column:version;type:varchar;size:8;default:2.1;" json:"version"`
	Recipe           string         `gorm:"column:recipe;type:longtext;" json:"recipe"`
}

func (s *ServerGd) TableName() string {
	return "servers_gd"
}

type ServerGdSmall struct {
	ID         int       `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"-"`
	SrvID      string    `gorm:"column:srvid;type:varchar;size:8;" json:"srvid"`
	SrvName    string    `gorm:"column:srvName;type:varchar;size:32;" json:"srv_name"`
	Plan       int       `gorm:"column:plan;type:tinyint;" json:"plan"`
	OwnerID    int       `gorm:"column:owner_id;type:int;" json:"owner_id"`
	UserCount  int       `gorm:"column:userCount;type:int;default:0;" json:"user_count"`
	LevelCount int       `gorm:"column:levelCount;type:int;default:0;" json:"level_count"`
	Icon       string    `gorm:"column:icon;type:varchar;size:16;default:gd_default.png;" json:"icon"`
	ExpireDate time.Time `gorm:"column:expireDate;type:datetime;default:CURRENT_TIMESTAMP;" json:"expire_date"`
}

type ServerGdReduced struct {
	ID               int    `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"-"`
	SrvID            string `gorm:"column:srvid;type:varchar;size:8;" json:"srvid"`
	Plan             int    `gorm:"column:plan;type:tinyint;" json:"plan"`
	SrvName          string `gorm:"column:srvName;type:varchar;size:32;" json:"srv_name"`
	Owner            string `gorm:"type:varchar;" json:"owner_id"`
	UserCount        int    `gorm:"column:userCount;type:int;default:0;" json:"user_count"`
	LevelCount       int    `gorm:"column:levelCount;type:int;default:0;" json:"level_count"`
	ClientAndroidURL string `gorm:"column:clientAndroidURL;type:varchar;size:255;" json:"client_android_url"`
	ClientIOSURL     string `gorm:"column:clientIOSURL;type:varchar;size:255;" json:"client_ios_url"`
	ClientWindowsURL string `gorm:"column:clientWindowsURL;type:varchar;size:255;" json:"client_windows_url"`
	ClientMacOSURL   string `gorm:"column:clientMacOSURL;type:varchar;size:255;" json:"client_macos_url"`
	Icon             string `gorm:"column:icon;type:varchar;size:16;default:gd_default.png;" json:"icon"`
	Description      string `gorm:"column:description;type:varchar;size:1000;default:Welcome to my GDPS!;" json:"description"`
	TextAlign        int    `gorm:"column:textAlign;type:tinyint;default:0;" json:"text_align"`
	Discord          string `gorm:"column:discord;type:varchar;size:128;" json:"discord"`
	Vk               string `gorm:"column:vk;type:varchar;size:128;" json:"vk"`
	Is22             bool   `gorm:"column:is22;type:tinyint;default:0;" json:"is_22"`
	IsCustomTextures bool   `gorm:"column:isCustomTextures;type:tinyint;default:0;" json:"is_custom_textures"`
}
