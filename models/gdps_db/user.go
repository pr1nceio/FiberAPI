package gdps_db

import (
	"database/sql"
	"time"
)

type User struct {
	UID                int32          `gorm:"primary_key;AUTO_INCREMENT;column:uid;type:int;" json:"uid"`
	Uname              string         `gorm:"column:uname;type:varchar;size:16;" json:"uname"`
	Passhash           string         `gorm:"column:passhash;type:varchar;size:128;" json:"passhash"`
	Email              string         `gorm:"column:email;type:varchar;size:256;" json:"email"`
	RoleID             int32          `gorm:"column:role_id;type:int;default:0;" json:"role_id"`
	Stars              int32          `gorm:"column:stars;type:int;default:0;" json:"stars"`
	Diamonds           int32          `gorm:"column:diamonds;type:int;default:0;" json:"diamonds"`
	Coins              int32          `gorm:"column:coins;type:int;default:0;" json:"coins"`
	Ucoins             int32          `gorm:"column:ucoins;type:int;default:0;" json:"ucoins"`
	Demons             int32          `gorm:"column:demons;type:int;default:0;" json:"demons"`
	Cpoints            int32          `gorm:"column:cpoints;type:int;default:0;" json:"cpoints"`
	Orbs               int32          `gorm:"column:orbs;type:int;default:0;" json:"orbs"`
	RegDate            time.Time      `gorm:"column:regDate;type:datetime;" json:"reg_date"`
	AccessDate         time.Time      `gorm:"column:accessDate;type:datetime;" json:"access_date"`
	LastIP             sql.NullString `gorm:"column:lastIP;type:varchar;size:64;default:Unknown;" json:"last_ip"`
	GameVer            sql.NullInt64  `gorm:"column:gameVer;type:int;default:20;" json:"game_ver"`
	LvlsCompleted      sql.NullInt64  `gorm:"column:lvlsCompleted;type:int;default:0;" json:"lvls_completed"`
	Special            int32          `gorm:"column:special;type:int;default:0;" json:"special"`
	ProtectMeta        string         `gorm:"column:protect_meta;type:text;size:4294967295;default:{'comm_time':0,'post_time':0,'msg_time':0};" json:"protect_meta"`
	ProtectLevelsToday int32          `gorm:"column:protect_levelsToday;type:int;default:0;" json:"protect_levels_today"`
	ProtectTodayStars  int32          `gorm:"column:protect_todayStars;type:int;default:0;" json:"protect_today_stars"`
	IsBanned           int32          `gorm:"column:isBanned;type:tinyint;default:0;" json:"is_banned"`
	Blacklist          string         `gorm:"column:blacklist;type:text;size:65535;" json:"blacklist"`
	FriendsCnt         int32          `gorm:"column:friends_cnt;type:int;default:0;" json:"friends_cnt"`
	FriendshipIds      string         `gorm:"column:friendship_ids;type:text;size:65535;" json:"friendship_ids"`
	IconType           int32          `gorm:"column:iconType;type:tinyint;default:0;" json:"icon_type"`
	Vessels            string         `gorm:"column:vessels;type:text;size:4294967295;default:{'clr_primary':0,'clr_secondary':0,'cube':0,'ship':0,'ball':0,'ufo':0,'wave':0,'robot':0,'spider':0,'swing':0,'jetpack':0,'trace':0,'death':0};" json:"vessels"`
	Chests             string         `gorm:"column:chests;type:text;size:4294967295;default:{'small_count':0,'big_count':0,'small_time':0,'big_time':0};" json:"chests"`
	Settings           string         `gorm:"column:settings;type:text;size:4294967295;default:{'frS':0,'cS':0,'mS':0,'youtube':'','twitch':'','twitter':''};" json:"settings"`
	Moons              int32          `gorm:"column:moons;type:int;default:0;" json:"moons"`
	Gjphash            string         `gorm:"column:gjphash;type:varchar;size:64;" json:"gjphash"`
}

func (u *User) TableName() string {
	return "users"
}
