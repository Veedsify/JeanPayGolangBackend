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

	emailJob := jobs.NewEmailJobClient()
	err = emailJob.EnqueueWelcomeEmail(user.Email, user.FirstName, "")
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

//	func VerifyUser(token string, email string) error {
//		// jwtService, err := libs.NewJWTServiceFromEnv()
//		// if err != nil {
//		// 	log.Fatal(err)
//		// }
//
//		// user, err := GetUserByEmail(email)
//		// if err != nil {
//		// 	return err
//		// }
//
//		// if user == nil {
//		// 	return errors.New("user not found")
//		// }
//
//		// _, err = jwtService.ValidateEmailVerificationToken(token)
//
//		// if err != nil {
//		// 	return err
//		// }
//
//		// database.DB.Model(&models.User{}).Where("email = ?", email).Update("is_verified", true)
//
//		return nil
//	}
func CreatePasswordReset(email string) (string, error) {
	user, err := GetUserByEmail(email)
	if err != nil {
		return "", errors.New("if your email exists in our system, you will receive a password reset link shortly")
	}

	if user == nil {
		return "", errors.New("user not found")
	}

	resetString := libs.GenerateRandomString(32)

	emailClient := jobs.NewEmailJobClient()
	defer emailClient.Close()

	emailClient.EnqueuePasswordResetEmail(email, resetString)
	redisclient := utils.NewRedisClient()
	cacheKey := fmt.Sprintf("password_reset:%s", resetString)
	err = utils.SetRedisKey(redisclient, cacheKey, email, time.Duration(15)*time.Minute)
	if err != nil {
		return "", errors.New("unable to create reset link")
	}
	return resetString, nil

}

func VerifyPasswordResetToken(token string) (string, error) {
	redisclient := utils.NewRedisClient()
	cacheKey := fmt.Sprintf("password_reset:%s", token)

	value, err := utils.GetRedisValue(redisclient, cacheKey)

	if err != nil {
		return "", err
	}

	if value == "" {
		return "", errors.New("no data found")
	}

	return value, nil

}

func ResetPassword(token string, password string) error {

	fmt.Printf("Resetting password for token %s and password %s", token, password)

	if password == "" || len(password) < 8 {
		return errors.New("password is required")
	}

	hashedPassword, err := libs.HashPassword(password)
	if err != nil {
		return err
	}

	email, err := VerifyPasswordResetToken(token)
	if err != nil {
		return err
	}

	if email == "" {
		return errors.New("invalid or expired token")
	}

	if err := database.DB.Model(&models.User{}).Where("email = ?", email).Update("password", hashedPassword).Error; err != nil {
		return err
	}

	redisclient := utils.NewRedisClient()
	cacheKey := fmt.Sprintf("password_reset:%s", token)
	utils.DeleteRedisKey(redisclient, cacheKey)

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

func RefreshToken(refreshToken string) (*libs.TokenPair, error) {
	jwtService, err := libs.NewJWTServiceFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	redisClient := utils.NewRedisClient()
	savedUserId, err := utils.GetRedisValue(redisClient, key)

	if err != nil {
		return &libs.TokenPair{}, errors.New("invalid refresh token")
	}

	userInfo, err := jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return &libs.TokenPair{}, errors.New("invalid refresh token")
	}

	var dbUser models.User

	err = database.DB.Where("id = ?", userInfo.ID).First(&dbUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &libs.TokenPair{}, errors.New("user not found")
		}
		return &libs.TokenPair{}, err
	}

	userID, err := libs.ConvertStringToUint(savedUserId)
	if err != nil {
		return &libs.TokenPair{}, errors.New("invalid refresh token")
	}

	if savedUserId == "" || userID != (dbUser.ID) {
		return &libs.TokenPair{}, errors.New("invalid refresh token")
	}

	if dbUser.IsBlocked {
		return &libs.TokenPair{}, errors.New("your account has been disabled, please contact support")
	}

	loggedInUser := &libs.UserInfo{
		ID:      dbUser.ID,
		UserID:  dbUser.UserID,
		Email:   dbUser.Email,
		IsAdmin: dbUser.IsAdmin,
	}

	// invalidate old refresh token
	utils.DeleteRedisKey(redisClient, key)

	newToken, err := jwtService.GenerateTokenPair(loggedInUser)
	if err != nil {
		return &libs.TokenPair{}, err
	}

	return newToken, nil
}
