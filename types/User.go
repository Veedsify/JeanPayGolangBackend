package types

import (
	"github.com/Veedsify/JeanPayGoBackend/database/models"
)

type RegisterUser struct {
	FirstName string             `json:"first_name" form:"first_name"`
	LastName  string             `json:"last_name" form:"last_name"`
	Email     string             `json:"email" form:"email"`
	Password  string             `json:"password" form:"password"`
	Country   models.UserCountry `json:"country" form:"country"`
}

type LoginUser struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

type UserWallet struct {
	ID               uint32 `json:"id"`
	UserID           uint32 `json:"user_id"`
	Currency         string
	TotalWithdrawals float64 `json:"total_withdrawals"`
	TotalConversions float64 `json:"total_conversions"`
	IsActive         bool    `json:"is_active" gorm:"default:true"`
}

type UserResponse struct {
	ID             int64
	UserID         uint32             `json:"user_id"`
	Email          string             `json:"email"`
	ProfilePicture string             `json:"profile_picture"`
	FirstName      string             `json:"first_name"`
	PhoneNumber    string             `json:"phone_number"`
	LastName       string             `json:"last_name"`
	Username       string             `json:"username"`
	IsAdmin        bool               `json:"is_admin"`
	IsBlocked      bool               `json:"is_blocked"`
	IsVerified     bool               `json:"is_verified"`
	Country        models.UserCountry `json:"country"`
	Setting        models.Setting     `json:"setting"`
}

type VerifyUser struct {
	Email string `json:"email" form:"email"`
}
