package models

import (
	"time"

	"gorm.io/gorm"
)

type UserCountry string

const (
	Nigeria UserCountry = "nigeria"
	Ghana   UserCountry = "ghana"
)

type User struct {
	gorm.Model
	FirstName          string        `json:"first_name" gorm:"not null"`
	LastName           string        `json:"last_name"`
	Email              string        `json:"email" gorm:"unique"`
	Username           string        `json:"username"`
	Password           string        `json:"password"`
	ProfilePicture     string        `json:"profile_picture" gorm:"default:'/images/defaults/user.jpg'"`
	PhoneNumber        string        `json:"phone_number"`
	IsAdmin            bool          `json:"is_admin"`
	IsVerified         bool          `json:"is_verified"`
	IsBlocked          bool          `json:"is_blocked"`
	UserID             uint32        `json:"user_id"`
	Country            UserCountry   `json:"country"`
	IsTwoFactorEnabled bool          `json:"is_two_factor_enabled"`
	UpdatedAt          time.Time     `json:"updated_at"`
	Setting            Setting       `json:"setting" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Transactions       []Transaction `json:"transactions" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Wallet             []Wallet      `json:"wallets" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Activity           []Activity    `json:"activities" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (User) TableName() string {
	return "users"
}
