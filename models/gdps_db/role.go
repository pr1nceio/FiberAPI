package gdps_db

import (
	"github.com/fruitspace/FiberAPI/models"
)

type Role struct {
	ID           int            `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	RoleName     string         `gorm:"column:roleName;type:varchar;size:64;default:Moderator;" json:"role_name"`
	CommentColor string         `gorm:"column:commentColor;type:varchar;size:11;default:0,0,255;" json:"comment_color"`
	ModLevel     int            `gorm:"column:modLevel;type:tinyint;default:1;" json:"mod_level"`
	Privs        models.JSONMap `gorm:"column:privs;type:json;size:65535;default:{'cRate':0,'cFeature':0,'cEpic':0,'cVerCoins':0,'cDaily':0,'cWeekly':0,'cDelete':0,'cLvlAccess':0,'aRateDemon':0,'aRateStars':0,'aReqMod':0,'dashboardMod':0,'dashboardBan':0,'dashboardCreatePack':0};" json:"privs"`
}

func (r *Role) TableName() string {
	return "roles"
}
