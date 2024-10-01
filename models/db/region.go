package db

import gorm "github.com/cradio/gormx"

type Region struct {
	gorm.Model
	Name             string
	Location         string
	Description      string
	PterodactylLocID uint
}

type RegionPublic struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Location    string `json:"location"`
	Description string `json:"description"`
}
