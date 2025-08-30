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
	"github.com/markbates/goth"
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
		IsVerified: false, // User starts unverified until they verify their email
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

	// Generate email verification token
	jwtService, err := libs.NewJWTServiceFromEnv()
	if err != nil {
		log.Printf("Error creating JWT service: %v", err)
		return errors.New("failed to create verification token")
	}

	userInfo := &libs.UserInfo{
		ID:      createUser.ID,
		UserID:  createUser.UserID,
		Email:   createUser.Email,
		IsAdmin: createUser.IsAdmin,
	}

	verificationToken, err := jwtService.GenerateEmailVerificationToken(userInfo)
	if err != nil {
		log.Printf("Error generating verification token: %v", err)
		return errors.New("failed to create verification token")
	}

	// Cache the verification token in Redis for 24 hours
	redisClient := utils.NewRedisClient()
	cacheKey := fmt.Sprintf("email_verification:%s", verificationToken)
	err = utils.SetRedisKey(redisClient, cacheKey, createUser.ID, time.Hour*24)
	if err != nil {
		log.Printf("Error caching verification token: %v", err)
		return errors.New("failed to cache verification token")
	}

	// Send welcome email with verification token
	emailJob := jobs.NewEmailJobClient()
	err = emailJob.EnqueueWelcomeEmail(user.Email, user.FirstName, verificationToken)
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

func VerifyEmailToken(token string) error {
	// Validate the JWT token
	jwtService, err := libs.NewJWTServiceFromEnv()
	if err != nil {
		return errors.New("failed to initialize JWT service")
	}

	claims, err := jwtService.ValidateEmailVerificationToken(token)
	if err != nil {
		return errors.New("invalid or expired verification token")
	}

	// Check if token exists in Redis cache
	redisClient := utils.NewRedisClient()
	cacheKey := fmt.Sprintf("email_verification:%s", token)
	cachedUserID, err := utils.GetRedisValue(redisClient, cacheKey)
	if err != nil || cachedUserID == "" {
		return errors.New("verification token has expired or already been used")
	}

	// Verify that the cached user ID matches the token claims
	userID, err := libs.ConvertStringToUint(cachedUserID)
	if err != nil || userID != claims.ID {
		return errors.New("invalid verification token")
	}

	// Update user verification status
	err = database.DB.Model(&models.User{}).Where("id = ?", claims.ID).Update("is_verified", true).Error
	if err != nil {
		return errors.New("failed to verify user account")
	}

	// Remove the token from cache after successful verification
	utils.DeleteRedisKey(redisClient, cacheKey)

	// Log activity
	activity := fmt.Sprintf(constants.EmailVerificationActivityLog, libs.FormatDate(time.Now()))
	jobs.NewActivityJobClient().EnqueueNewActivity(claims.ID, activity)

	return nil
}

func ResendEmailVerification(email string) error {
	// Get user by email from database
	var user models.User
	err := database.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return errors.New("database error")
	}

	if user.IsVerified {
		return errors.New("user is already verified")
	}

	// Generate new email verification token
	jwtService, err := libs.NewJWTServiceFromEnv()
	if err != nil {
		log.Printf("Error creating JWT service: %v", err)
		return errors.New("failed to create verification token")
	}

	userInfo := &libs.UserInfo{
		ID:      user.ID,
		UserID:  user.UserID,
		Email:   user.Email,
		IsAdmin: user.IsAdmin,
	}

	verificationToken, err := jwtService.GenerateEmailVerificationToken(userInfo)
	if err != nil {
		log.Printf("Error generating verification token: %v", err)
		return errors.New("failed to create verification token")
	}

	// Cache the verification token in Redis for 24 hours
	redisClient := utils.NewRedisClient()
	cacheKey := fmt.Sprintf("email_verification:%s", verificationToken)
	err = utils.SetRedisKey(redisClient, cacheKey, user.ID, time.Hour*24)
	if err != nil {
		log.Printf("Error caching verification token: %v", err)
		return errors.New("failed to cache verification token")
	}

	// Send verification email
	emailJob := jobs.NewEmailJobClient()
	err = emailJob.EnqueueWelcomeEmail(user.Email, user.FirstName, verificationToken)
	if err != nil {
		fmt.Printf("Error creating verification email job: %v\n", err)
		return errors.New("failed to send verification email")
	}

	return nil
}
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

func GoogleAuthLogin(user goth.User) (*libs.TokenPair, string, error) {
	if user.Email == "" {
		return &libs.TokenPair{}, "login", errors.New("unable to get email from Google account")
	}

	var dbUser models.User
	err := database.DB.Where("email = ?", user.Email).First(&dbUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ngnId, ghsId := libs.GenerateUniqueWalletId()
			createUser := models.User{
				Email:         user.Email,
				FirstName:     user.FirstName,
				LastName:      user.LastName,
				IsAdmin:       false,
				IsVerified:    true,
				IsBlocked:     false,
				GoogleCloudId: user.UserID,
				UserID:        uuid.New().ID(),
				Country:       "NGN",
				Setting: models.Setting{
					DefaultCurrency: models.DefaultCurrency("NGN"),
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
				}}

			if err := database.DB.Create(&createUser).Error; err != nil {
				return &libs.TokenPair{}, "login", errors.New("sorry this account already exists")
			}
			dbUser = createUser
		} else {
			return &libs.TokenPair{}, "login", err
		}
	}

	if dbUser.IsBlocked {
		return &libs.TokenPair{}, "login", errors.New("your account has been disabled, please contact support")
	}

	if dbUser.GoogleCloudId != user.UserID {
		return &libs.TokenPair{}, "login", errors.New("this account was not registered with Google, please use your email and password to login")
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
		return &libs.TokenPair{}, "login", err
	}

	token, err := jwtService.GenerateTokenPair(loggedInUser)

	if err != nil {
		return &libs.TokenPair{}, "login", err
	}

	activity := fmt.Sprintf(constants.NewLoginActivityLog, libs.FormatDate(time.Now()))
	jobs.NewActivityJobClient().EnqueueNewActivity(dbUser.ID, activity)
	return token, "login", nil
}
