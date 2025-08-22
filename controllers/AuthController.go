package controllers

import (
	"fmt"
	"net/http"

	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/gin-gonic/gin"
)

func RegisterUserEndpoint(c *gin.Context) {
	var user types.RegisterUser
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error(), "error": true})
		return
	}

	if err := services.RegisterUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error(), "error": true})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"error": false, "message": "User registered successfully"})
}

func LoginUserEndpoint(c *gin.Context) {
	var user types.LoginUser
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error(), "error": true})
		return
	}

	token, action, err := services.LoginUser(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error(), "error": true, "action": action})
		return
	}

	if token.IsAdmin {
		c.SetCookie("admin_token", token.AccessToken, 3600, "/", "", false, true)
		c.SetCookie("refresh_token", token.RefreshToken, 3600, "/", "", false, true)
		c.Header("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"error": false,
		})
		return
	}
	c.SetCookie("token", token.AccessToken, 3600, "/", "", false, true)
	c.SetCookie("refresh_token", token.RefreshToken, 3600, "/", "", false, true)
	c.Header("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"error": false,
	})

}

//func VerifyUserEndpoint(c *gin.Context) {
//	var email = c.Param("email")
//	var token = c.Param("token")
//
//	if email == "" || token == "" {
//		c.JSON(http.StatusBadRequest, gin.H{"message": "email and token are required", "error": true})
//		return
//	}
//
//	if err := services.VerifyUser(token, email); err != nil {
//		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error(), "error": true})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{"message": "User verified successfully", "error": false})
//}
//
//func PasswordResetEndpoint(c *gin.Context) {
//	email := c.PostForm("email")
//	if email == "" {
//		c.JSON(http.StatusBadRequest, gin.H{"message": "email is required", "error": true})
//		return
//	}
//
//	resetToken, err := services.PasswordReset(email)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error(), "error": true})
//		return
//	}
//
//	if resetToken == "" {
//		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to generate reset token", "error": true})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"message": "Password reset email sent successfully",
//		"error":   false,
//	})
//
//}
//
//func ResetPasswordTokenVerifyEndpoint(c *gin.Context) {
//	token := c.Query("token")
//	if token == "" {
//		c.JSON(http.StatusBadRequest, gin.H{"message": "token is required", "error": true})
//		return
//	}
//
//	if err := services.VerifyPasswordResetToken(token); err != nil {
//		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error(), "error": true})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"message": "Token verified successfully",
//		"error":   false,
//	})
//
//}
//
//func ResetPasswordEndpoint(c *gin.Context) {
//	password := c.PostForm("password")
//	token := c.Query("token")
//	if password == "" || len(password) < 8 {
//		c.JSON(http.StatusBadRequest, gin.H{"message": "password is required", "error": true})
//		return
//	}
//
//	if err := services.ResetPassword(token, password); err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error(), "error": true})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"message": "Password reset successfully",
//		"error":   false,
//	})
//
//}

func VerifyOtpEndpoint(c *gin.Context) {
	var otp struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBind(&otp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error(), "error": true})
		return
	}

	if otp.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "OTP and User ID are required", "error": true})
		return
	}

	token, code, err := services.VerifyOtp(otp.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error(), "error": true, "action": code})
		return
	}

	if token.IsAdmin {
		c.SetCookie("admin_token", token.AccessToken, 3600, "/", "", false, true)
		c.SetCookie("refresh_token", token.RefreshToken, 3600, "/", "", false, true)
		c.Header("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"error": false,
		})
		return
	}
	c.SetCookie("token", token.AccessToken, 3600, "/", "", false, true)
	c.SetCookie("refresh_token", token.RefreshToken, 3600, "/", "", false, true)
	c.Header("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"error": false,
	})
}
