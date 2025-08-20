package models

import (
	"gorm.io/gorm"
)

type AdminLog struct {
	gorm.Model
	AdminID   uint32 `json:"admin_id" gorm:"not null"`
	Action    string `json:"action" gorm:"not null"`
	Target    string `json:"target" gorm:"not null"` // e.g., 'transaction', 'user', 'rate'
	TargetID  string `json:"target_id" gorm:"not null"`
	Details   string `json:"details" gorm:"type:string"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

func (AdminLog) TableName() string {
	return "admin_logs"
}
