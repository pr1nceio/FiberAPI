package gdps_db

import (
	"time"
)

type Level struct {
	ID                   int       `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	Name                 string    `gorm:"column:name;type:varchar;size:32;default:Unnamed;" json:"name"`
	Description          string    `gorm:"column:description;type:varchar;size:256;" json:"description"`
	UID                  int       `gorm:"column:uid;type:int;" json:"uid"`
	Password             string    `gorm:"column:password;type:varchar;size:8;" json:"password"`
	Version              int       `gorm:"column:version;type:tinyint;default:1;" json:"version"`
	Length               int       `gorm:"column:length;type:tinyint;default:0;" json:"length"`
	Difficulty           int       `gorm:"column:difficulty;type:tinyint;default:0;" json:"difficulty"`
	DemonDifficulty      int       `gorm:"column:demonDifficulty;type:tinyint;default:-1;" json:"demon_difficulty"`
	SuggestDifficulty    float64   `gorm:"column:suggestDifficulty;type:float;default:0.0;" json:"suggest_difficulty"`
	SuggestDifficultyCnt int       `gorm:"column:suggestDifficultyCnt;type:int;default:0;" json:"suggest_difficulty_cnt"`
	TrackID              int       `gorm:"column:track_id;type:mediumint;default:0;" json:"track_id"`
	SongID               int       `gorm:"column:song_id;type:mediumint;default:0;" json:"song_id"`
	VersionGame          int       `gorm:"column:versionGame;type:tinyint;" json:"version_game"`
	VersionBinary        int       `gorm:"column:versionBinary;type:tinyint;" json:"version_binary"`
	StringExtra          string    `gorm:"column:stringExtra;type:text;size:16777215;" json:"string_extra"`
	StringLevel          string    `gorm:"column:stringLevel;type:text;size:4294967295;" json:"string_level"`
	StringLevelInfo      string    `gorm:"column:stringLevelInfo;type:text;size:16777215;" json:"string_level_info"`
	OriginalID           int       `gorm:"column:original_id;type:int;default:0;" json:"original_id"`
	Objects              uint      `gorm:"column:objects;type:uint;" json:"objects"`
	StarsRequested       int       `gorm:"column:starsRequested;type:tinyint;" json:"stars_requested"`
	StarsGot             int       `gorm:"column:starsGot;type:tinyint;default:0;" json:"stars_got"`
	Ucoins               int       `gorm:"column:ucoins;type:tinyint;" json:"ucoins"`
	Coins                int       `gorm:"column:coins;type:tinyint;default:0;" json:"coins"`
	Downloads            uint      `gorm:"column:downloads;type:uint;default:0;" json:"downloads"`
	Likes                int       `gorm:"column:likes;type:int;default:0;" json:"likes"`
	Reports              uint      `gorm:"column:reports;type:uint;default:0;" json:"reports"`
	Collab               string    `gorm:"column:collab;type:text;size:65535;" json:"collab"`
	Is2p                 int       `gorm:"column:is2p;type:tinyint;default:0;" json:"is_2_p"`
	IsVerified           int       `gorm:"column:isVerified;type:tinyint;default:0;" json:"is_verified"`
	IsFeatured           int       `gorm:"column:isFeatured;type:tinyint;default:0;" json:"is_featured"`
	IsHall               int       `gorm:"column:isHall;type:tinyint;default:0;" json:"is_hall"`
	IsEpic               int       `gorm:"column:isEpic;type:tinyint;default:0;" json:"is_epic"`
	IsUnlisted           int       `gorm:"column:isUnlisted;type:tinyint;default:0;" json:"is_unlisted"`
	IsLDM                int       `gorm:"column:isLDM;type:tinyint;default:0;" json:"is_ldm"`
	UploadDate           time.Time `gorm:"column:uploadDate;type:datetime;" json:"upload_date"`
	UpdateDate           time.Time `gorm:"column:updateDate;type:datetime;" json:"update_date"`
	StringSettings       string    `gorm:"column:stringSettings;type:text;size:16777215;" json:"string_settings"`
}

func (l *Level) TableName() string {
	return "levels"
}
