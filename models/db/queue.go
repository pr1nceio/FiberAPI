package db

import (
	"database/sql"
	"time"
)

type Queue struct {
	ID        int            `gorm:"primary_key;AUTO_INCREMENT;column:id;type:int;" json:"id"`
	Type      string         `gorm:"column:type;type:varchar;size:8;" json:"type"`
	SrvID     string         `gorm:"column:srvId;type:varchar;size:64;" json:"srv_id"`
	Worker    string         `gorm:"column:worker;type:varchar;size:128;" json:"worker"`
	UpdatedAt time.Time      `gorm:"column:startTime;type:datetime;default:CURRENT_TIMESTAMP;" json:"start_time"`
	Data      sql.NullString `gorm:"column:data;type:json;" json:"data"`
}

func (q *Queue) TableName() string {
	return "queue"
}
