package db

import gorm "github.com/cradio/gormx"

type Particle struct {
	gorm.Model
	Name        string `json:"name"`
	Author      string `json:"author"`
	UID         uint   `json:"uid"`
	Arch        string `json:"arch"`
	LayerID     string `json:"-"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Recipe      string `json:"recipe"`
	Size        uint   `gorm:"default:0" json:"size"` // bytes
	IsPrivate   bool   `gorm:"default:0" json:"is_private"`
	IsUnlisted  bool   `gorm:"default:0" json:"is_unlisted"`
	Downloads   uint   `gorm:"default:0" json:"downloads"`
	Likes       uint   `gorm:"default:0" json:"likes"`
	IsOfficial  bool   `gorm:"default:0" json:"is_official"`
}

func (p *Particle) TableName() string {
	return "particles"
}
