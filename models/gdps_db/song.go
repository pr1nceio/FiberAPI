package gdps_db

type Song struct {
	ID        int32   `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	AuthorID  int32   `gorm:"column:author_id;type:int;default:0;" json:"author_id"`
	Name      string  `gorm:"column:name;type:varchar;size:128;default:Unnamed;" json:"name"`
	Artist    string  `gorm:"column:artist;type:varchar;size:128;default:Unknown;" json:"artist"`
	Size      float32 `gorm:"column:size;type:float;" json:"size"`
	URL       string  `gorm:"column:url;type:varchar;size:1024;" json:"url"`
	IsBanned  int32   `gorm:"column:isBanned;type:tinyint;default:0;" json:"is_banned"`
	Downloads int32   `gorm:"column:downloads;type:int;default:0;" json:"downloads"`
}

func (s *Song) TableName() string {
	return "songs"
}
