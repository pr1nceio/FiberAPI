package db

import (
	"time"
)

type User struct {
	UID          int       `gorm:"primary_key;AUTO_INCREMENT;column:uid;type:int;" json:"uid"`
	Uname        string    `gorm:"column:uname;type:varchar;size:32;" json:"uname"`
	Name         string    `gorm:"column:name;type:varchar;size:128;" json:"name"`
	Surname      string    `gorm:"column:surname;type:varchar;size:128;" json:"surname"`
	Email        string    `gorm:"column:email;type:varchar;size:255;" json:"email"`
	ProfilePic   string    `gorm:"column:profilePic;type:varchar;size:64;default:default.jpg;" json:"profile_pic"`
	PassHash     string    `gorm:"column:passHash;type:varchar;size:255;" json:"-"`
	TotpSecret   string    `gorm:"column:totpSecret;type:varchar;size:255;" json:"-"`
	VkID         string    `gorm:"column:vk_id;type:bigint;default:0;" json:"vk_id"`
	VkToken      string    `gorm:"column:vk_token;type:varchar;size:255;" json:"-"`
	DiscordID    string    `gorm:"column:discord_id;type:bigint;default:0;" json:"discord_id"`
	DiscordToken string    `gorm:"column:discord_token;type:varchar;size:255;" json:"-"`
	Balance      float64   `gorm:"column:balance;type:double;default:0.00;" json:"balance"`
	BalanceHash  string    `gorm:"column:balanceHash;type:varchar;size:255;" json:"-"`
	Affiliate    int       `gorm:"column:affiliate;type:int;default:0;" json:"affiliate"`
	Reflink      string    `gorm:"column:reflink;type:varchar;size:255;" json:"reflink"`
	CreatedAt    time.Time `gorm:"column:regDate;type:datetime;default:CURRENT_TIMESTAMP;" json:"-"`
	UpdatedAt    time.Time `gorm:"column:accessDate;type:datetime;default:CURRENT_TIMESTAMP;" json:"-"`
	LastIP       string    `gorm:"column:lastIP;type:varchar;size:255;default:Unknown;" json:"-"`
	Country      string    `gorm:"column:country;type:varchar;size:8;default:Unknown;" json:"-"`
	City         string    `gorm:"column:city;type:varchar;size:127;default:Unknown;" json:"-"`
	Provider     string    `gorm:"column:provider;type:varchar;size:255;default:Unknown;" json:"-"`
	IsActivated  bool      `gorm:"column:isActivated;type:tinyint;default:0;" json:"-"`
	IsBanned     bool      `gorm:"column:isBanned;type:tinyint;default:0;" json:"-"`
	IsPartner    bool      `gorm:"column:isPartner;type:tinyint;default:0;" json:"is_partner"`
	IsAdmin      bool      `gorm:"column:isAdmin;type:tinyint;default:0;" json:"is_admin"`
	Is2FA        bool      `gorm:"column:is2FA;type:tinyint;default:0;" json:"is_2fa"`
}

func (u *User) TableName() string {
	return "users"
}
