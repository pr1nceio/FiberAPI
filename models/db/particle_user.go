package db

import gorm "github.com/cradio/gormx"

type ParticleUser struct {
	gorm.Model
	Username       string `json:"username"`
	Token          string `json:"token"`
	MaxAllowedSize uint   `json:"max_allowed_size" gorm:"default:0"` //bytes
	IsAdmin        bool   `json:"is_admin"`
}

func (u *ParticleUser) TableName() string {
	return "particle_users"
}
