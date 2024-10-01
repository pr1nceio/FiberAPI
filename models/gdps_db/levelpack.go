package gdps_db

type LevelPack struct {
	ID             int    `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	PackType       int    `gorm:"column:packType;type:tinyint;" json:"pack_type"`
	PackName       string `gorm:"column:packName;type:varchar(256);size:256;" json:"pack_name"`
	Levels         string `gorm:"column:levels;type:varchar(512);size:512;" json:"levels"`
	PackStars      int    `gorm:"column:packStars;type:tinyint;default:0;" json:"pack_stars"`
	PackCoins      int    `gorm:"column:packCoins;type:tinyint;default:0;" json:"pack_coins"`
	PackDifficulty int    `gorm:"column:packDifficulty;type:tinyint;" json:"pack_difficulty"`
	PackColor      string `gorm:"column:packColor;type:varchar(11);size:11;" json:"pack_color"`
}

func (l *LevelPack) TableName() string {
	return "levelpacks"
}
