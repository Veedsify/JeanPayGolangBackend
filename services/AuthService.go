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

	if dbUser.IsTwoFactorEnabled {
		enabled := true

		// jwtService, err := libs.NewJWTServiceFromEnv()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// emailToken, err := jwtService.GenerateEmailVerificationToken(dbUser.UserId, dbUser.Email)
		// if err != nil {
		// 	return &libs.TokenPair{}, err
		// }
		// htmlEmail := new(strings.Builder)
		// template.Must(template.New("email").Parse(`
		// 	<div>
		// 	    {{.FirstName}} {{.LastName}}<br>
		// 	    {{.Email}}<br>
		// 	    Here is your email token:
		// 	    <a href="http://localhost:8080/verify-email/{{.EmailToken}}">HERE</a>
		// 	</div>
		// 	`)).Execute(htmlEmail, map[string]string{
		// 	"FirstName":  dbUser.FirstName,
		// 	"LastName":   dbUser.LastName,
		// 	"Email":      dbUser.Email,
		// 	"EmailToken": emailToken,
		// })

		// verifyUser, err := NewEmailServiceFromEnv()
		// if err != nil {
		// 	return &libs.TokenPair{}, err
		// }

		// if err := verifyUser.SendHTMLEmail([]string{dbUser.Email}, "Verify Your Account", htmlEmail.String()); err != nil {
		// 	return &libs.TokenPair{}, err
		// }

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
