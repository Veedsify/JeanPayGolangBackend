package models

import "gorm.io/gorm"

type Contact struct {
	*gorm.Model
	Email       string `json:"email" gorm:"not null"`
	FullName    string `json:"full_name" gorm:"not null"`
	PhoneNumber string `json:"phone_number" gorm:"not null"`
	Category    string `json:"category" gorm:"not null"`
	Subject     string `json:"subject" gorm:"not null"`
	Message     string `json:"message" gorm:"not null"`
	Delivered   bool   `json:"delivered" gorm:"default:false"`
}

func (Contact) TableName() string {
	return "contacts"
}
