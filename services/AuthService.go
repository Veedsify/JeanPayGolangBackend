package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/constants"
	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/jobs"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/Veedsify/JeanPayGoBackend/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func RegisterUser(user types.RegisterUser) error {
	uniqUUid := uuid.New().ID()

	hashedPassword, err := libs.HashPassword(user.Password)
	if err != nil {
		return err
	}

	ngnId, ghsId := libs.GenerateUniqueWalletId()

	createUser := models.User{
		Email:      user.Email,
		Password:   hashedPassword,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Country:    user.Country,
		IsAdmin:    false,
		IsVerified: true,
		UserID:     uniqUUid,
		Setting: models.Setting{
			DefaultCurrency: models.DefaultCurrency(libs.GetDefaultCurrency(string(user.Country))),
		},
		Wallet: []models.Wallet{
			{
				Currency: "NGN",
				Balance:  0,
				WalletID: ngnId,
			},
			{
				Currency: "GHS",
				Balance:  0,
				WalletID: ghsId,
			},
		},
	}

	if err := database.DB.Create(&createUser).Error; err != nil {
		return errors.New("sorry this account already exists")
	}

	jwtService, err := libs.NewJWTServiceFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	token, err := jwtService.GenerateEmailVerificationToken(createUser.ID, createUser.UserID, createUser.Email)

	if err != nil {
		return err
	}

	emailJob := jobs.NewEmailJobClient()
	err = emailJob.EnqueueWelcomeEmail(user.Email, user.FirstName, token)
	if err != nil {
		fmt.Printf("Error creating welcome email job: %v\n", err)
	}
	return nil
}

func LoginUser(user types.LoginUser) (*libs.TokenPair, string, error) {
	var dbUser models.User
	err := database.DB.Where("email = ?", user.Email).First(&dbUser).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &libs.TokenPair{}, "login", errors.New("invalid Email Address Or Password")
		}
		return &libs.TokenPair{}, "login", err
	}

	if err := libs.ComparePassword(dbUser.Password, user.Password); err != nil {
		return &libs.TokenPair{}, "login", errors.New("invalid Email Address Or Password")
	}

	if !dbUser.IsVerified {
		return &libs.TokenPair{}, "verify", errors.New("your Account Is Not Verified")
	}

	if dbUser.IsBlocked {
		return &libs.TokenPair{}, "login", errors.New("your account has been disabled, please contact support")
	}

	if dbUser.IsTwoFactorEnabled {
		enabled := true

		randomVerificationCode := libs.GenerateOTP(6)

		// deterministic hash for Redis key
		hashedKey := libs.SHA256(randomVerificationCode)

		redisClient := utils.NewRedisClient()
		cacheKey := fmt.Sprintf("two_factor:%s", hashedKey)

		// store userID, expire in 10 min
		utils.SetRedisKey(redisClient, cacheKey, dbUser.ID, time.Minute*10)

		// send raw code to user
		emailClient := jobs.NewEmailJobClient()
		emailClient.EnqueueTwoFactorEmail(dbUser.Email, dbUser.FirstName, randomVerificationCode)

		return &libs.TokenPair{
			IsTwoFactorEnabled: &enabled,
			AccessToken:        "",
			RefreshToken:       "",
		}, "login", nil
	}

	loggedInUser := &libs.UserInfo{
		ID:      dbUser.ID,
		UserID:  dbUser.UserID,
		Email:   dbUser.Email,
		IsAdmin: dbUser.IsAdmin,
	}

	jwtService, err := libs.NewJWTServiceFromEnv()

	if err != nil {
		log.Fatal(err)
	}

	token, err := jwtService.GenerateTokenPair(loggedInUser)
	if err != nil {
		return &libs.TokenPair{}, "login", err
	}
	activity := fmt.Sprintf(constants.NewLoginActivityLog, libs.FormatDate(time.Now()))
	jobs.NewActivityJobClient().EnqueueNewActivity(dbUser.ID, activity)
	return token, "login", nil
}

func VerifyUser(token string, email string) error {
	jwtService, err := libs.NewJWTServiceFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	user, err := GetUserByEmail(email)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user not found")
	}

	_, err = jwtService.ValidateEmailVerificationToken(token)

	if err != nil {
		return err
	}

	database.DB.Model(&models.User{}).Where("email = ?", email).Update("is_verified", true)

	return nil
}

func PasswordReset(email string) (string, error) {
	user, err := GetUserByEmail(email)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", errors.New("user not found")
	}

	jwtService, err := libs.NewJWTServiceFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	resetToken, err := jwtService.GeneratePasswordResetToken(user.ID, user.UserID, user.Email)
	if err != nil {
		return "", err
	}

	verifyUser, err := NewEmailServiceFromEnv()
	if err != nil {
		return "", err
	}

	if err := verifyUser.SendPasswordResetEmail(email, resetToken); err != nil {
		return "", err
	}

	return resetToken, nil

}

func VerifyPasswordResetToken(token string) error {
	jwtService, err := libs.NewJWTServiceFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	_, err = jwtService.ValidatePasswordResetToken(token)

	if err != nil {
		return err
	}

	return nil

}

func ResetPassword(token string, password string) error {
	if password == "" || len(password) < 8 {
		return errors.New("password is required")
	}

	hashedPassword, err := libs.HashPassword(password)
	if err != nil {
		return err
	}

	jwtService, err := libs.NewJWTServiceFromEnv()
	if err != nil {
		return err
	}

	claims, err := jwtService.ValidatePasswordResetToken(token)
	if err != nil {
		return err
	}

	userId := claims.UserID
	database.DB.Model(&models.User{}).Where("user_id = ?", userId).Update("password", hashedPassword)

	return nil

}

func VerifyOtp(code string) (*libs.TokenPair, string, error) {
	hashedKey := libs.SHA256(code)

	redisClient := utils.NewRedisClient()
	cacheKey := fmt.Sprintf("two_factor:%s", hashedKey)

	// get userID from cache
	cachedUserId, err := utils.GetRedisValue(redisClient, cacheKey)

	if err != nil {
		return &libs.TokenPair{}, "login", errors.New("invalid or expired code")
	}

	if cachedUserId == "" {
		return &libs.TokenPair{}, "login", errors.New("invalid or expired code")
	}

	// success: delete OTP after use
	utils.DeleteRedisKey(redisClient, cacheKey)

	jwtService, err := libs.NewJWTServiceFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	var dbUser models.User
	if err := database.DB.Where("id = ?", cachedUserId).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &libs.TokenPair{}, "login", errors.New("invalid Email Address Or Password")
		}
	}

	loggedInUser := &libs.UserInfo{
		ID:      dbUser.ID,
		UserID:  dbUser.UserID,
		Email:   dbUser.Email,
		IsAdmin: dbUser.IsAdmin,
	}

	token, err := jwtService.GenerateTokenPair(loggedInUser)
	if err != nil {
		return &libs.TokenPair{}, "login", err
	}
	activity := fmt.Sprintf(constants.NewLoginActivityLog, libs.FormatDate(time.Now()))
	jobs.NewActivityJobClient().EnqueueNewActivity(dbUser.ID, activity)
	return token, "login", nil
}
