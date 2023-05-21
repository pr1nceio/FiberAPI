package gdps_db

type LevelPack struct {
	ID             int32  `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	PackType       int32  `gorm:"column:packType;type:tinyint;" json:"pack_type"`
	PackName       string `gorm:"column:packName;type:varchar;size:256;" json:"pack_name"`
	Levels         string `gorm:"column:levels;type:varchar;size:512;" json:"levels"`
	PackStars      int32  `gorm:"column:packStars;type:tinyint;default:0;" json:"pack_stars"`
	PackCoins      int32  `gorm:"column:packCoins;type:tinyint;default:0;" json:"pack_coins"`
	PackDifficulty int32  `gorm:"column:packDifficulty;type:tinyint;" json:"pack_difficulty"`
	PackColor      string `gorm:"column:packColor;type:varchar;size:11;" json:"pack_color"`
}

func (l *LevelPack) TableName() string {
	return "levelpacks"
}
