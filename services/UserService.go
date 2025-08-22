package services

import (
	"errors"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"gorm.io/gorm"
)

func FetchUser(UserId uint) (types.UserResponse, error) {
	if UserId == 0 {
		return types.UserResponse{}, errors.New("user Id Is Required")
	}

	var userData models.User

	if err := database.DB.Preload("Setting").Where("id = ?", UserId).First(&userData).Error; err != nil {
		return types.UserResponse{}, errors.New("user Data Not Found")
	}

	response := types.UserResponse{
		ID:             userData.ID,
		UserID:         userData.UserID,
		Email:          userData.Email,
		IsAdmin:        userData.IsAdmin,
		PhoneNumber:    userData.PhoneNumber,
		ProfilePicture: userData.ProfilePicture,
		FirstName:      userData.FirstName,
		LastName:       userData.LastName,
		Username:       userData.Username,
		Country:        userData.Country,
		IsBlocked:      userData.IsBlocked,
		IsVerified:     userData.IsVerified,
		Setting:        userData.Setting,
	}

	return response, nil
}

func GetUserById(id uint) (*libs.UserInfo, error) {
	var user models.User

	err := database.DB.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil user and nil error for not found
		}
		return nil, err
	}

	return &libs.UserInfo{
		ID:      user.ID,
		UserID:  user.UserID,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
	}, nil
}

func GetUserByEmail(email string) (*libs.UserInfo, error) {
	var user models.User
	err := database.DB.Where("email = ?", email).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil user and nil error for not found
		}
		return nil, err
	}

	return &libs.UserInfo{
		UserID:  user.UserID,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
	}, nil
}
