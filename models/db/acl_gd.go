package db

type ACLGd struct {
	ID               string `gorm:"primary_key;column:id;type:varchar;size:64;" json:"id"`
	UID              int    `gorm:"column:uid;type:int;" json:"uid"`
	TargetID         string `gorm:"column:target_id;type:varchar;size:64;" json:"target_id"`
	CanAccess        bool   `gorm:"column:canAccess;type:tinyint;default:0;" json:"can_access"`
	CanModerateMusic bool   `gorm:"column:canModerateMusic;type:tinyint;default:0;" json:"can_moderate_music"`
	CanViewPlayers   bool   `gorm:"column:canViewPlayers;type:tinyint;default:0;" json:"can_view_players"`
	CanBanPlayers    bool   `gorm:"column:canBanPlayers;type:tinyint;default:0;" json:"can_ban_players"`
	CanEditRoles     bool   `gorm:"column:canEditRoles;type:tinyint;default:0;" json:"can_edit_roles"`
	CanEditChests    bool   `gorm:"column:canEditChests;type:tinyint;default:0;" json:"can_edit_chests"`
	CanEditEvents    bool   `gorm:"column:canEditEvents;type:tinyint;default:0;" json:"can_edit_events"`
	CanEditMappacks  bool   `gorm:"column:canEditMappacks;type:tinyint;default:0;" json:"can_edit_mappacks"`
	CanEditGauntlets bool   `gorm:"column:canEditGauntlets;type:tinyint;default:0;" json:"can_edit_gauntlets"`
	CanViewLogs      bool   `gorm:"column:canViewLogs;type:tinyint;default:0;" json:"can_view_logs"`
	CanClearLogs     bool   `gorm:"column:canClearLogs;type:tinyint;default:0;" json:"can_clear_logs"`
	CanEditServer    bool   `gorm:"column:canEditServer;type:tinyint;default:0;" json:"can_edit_server"`
	CanQueryBuild    bool   `gorm:"column:canQueryBuild;type:tinyint;default:0;" json:"can_query_build"`
	CanEditShop      bool   `gorm:"column:canEditShop;type:tinyint;default:0;" json:"can_edit_shop"`
	IsAdmin          bool   `gorm:"column:isAdmin;type:tinyint;default:0;" json:"is_admin"`
}

func (a *ACLGd) TableName() string {
	return "acl_gd"
}
