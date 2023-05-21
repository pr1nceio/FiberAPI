package db

type ACLGd struct {
	ID               string `gorm:"primary_key;column:id;type:varchar;size:64;" json:"id"`
	UID              int    `gorm:"column:uid;type:int;" json:"uid"`
	TargetID         string `gorm:"column:target_id;type:varchar;size:64;" json:"target_id"`
	CanAccess        int    `gorm:"column:canAccess;type:tinyint;default:0;" json:"can_access"`
	CanModerateMusic int    `gorm:"column:canModerateMusic;type:tinyint;default:0;" json:"can_moderate_music"`
	CanViewPlayers   int    `gorm:"column:canViewPlayers;type:tinyint;default:0;" json:"can_view_players"`
	CanBanPlayers    int    `gorm:"column:canBanPlayers;type:tinyint;default:0;" json:"can_ban_players"`
	CanEditRoles     int    `gorm:"column:canEditRoles;type:tinyint;default:0;" json:"can_edit_roles"`
	CanEditChests    int    `gorm:"column:canEditChests;type:tinyint;default:0;" json:"can_edit_chests"`
	CanEditEvents    int    `gorm:"column:canEditEvents;type:tinyint;default:0;" json:"can_edit_events"`
	CanEditMappacks  int    `gorm:"column:canEditMappacks;type:tinyint;default:0;" json:"can_edit_mappacks"`
	CanEditGauntlets int    `gorm:"column:canEditGauntlets;type:tinyint;default:0;" json:"can_edit_gauntlets"`
	CanViewLogs      int    `gorm:"column:canViewLogs;type:tinyint;default:0;" json:"can_view_logs"`
	CanClearLogs     int    `gorm:"column:canClearLogs;type:tinyint;default:0;" json:"can_clear_logs"`
	CanEditServer    int    `gorm:"column:canEditServer;type:tinyint;default:0;" json:"can_edit_server"`
	CanQueryBuild    int    `gorm:"column:canQueryBuild;type:tinyint;default:0;" json:"can_query_build"`
	CanEditShop      int    `gorm:"column:canEditShop;type:tinyint;default:0;" json:"can_edit_shop"`
	IsAdmin          int    `gorm:"column:isAdmin;type:tinyint;default:0;" json:"is_admin"`
}

func (a *ACLGd) TableName() string {
	return "acl_gd"
}
