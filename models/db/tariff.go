package db

import gorm "github.com/cradio/gormx"

type Tariff struct {
	gorm.Model  `json:"-"`
	Name        string  `json:"name"`
	IsDynamic   bool    `json:"is_dynamic"`
	Short       string  `json:"short"`
	Description string  `json:"description"`
	CPU         float64 `json:"cpu"`
	MinRamMB    uint    `json:"min_ram_mb"`
	MaxRamMB    uint    `json:"max_ram_mb"`
	DiskGB      uint    `json:"disk_gb"`
}
