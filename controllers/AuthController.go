package controllers

import (
	"fmt"
	"net/http"

	"github.com/Veedsify/JeanPayGoBackend/libs"
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

	if token.IsTwoFactorEnabled != nil && *token.IsTwoFactorEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "Two-factor authentication required",
			"error":   false,
			"action":  action,
			"token":   token,
		})
		return
	}

	if token.IsAdmin {
		if err := libs.SetCookie(c, "admin_token", token.AccessToken, 3600, "/"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set cookie", "error": err.Error()})
			return
		}
		if err := libs.SetCookie(c, "refresh_token", token.RefreshToken, 3600, "/"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set cookie", "error": err.Error()})
			return
		}
		c.Header("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"error": false,
		})
		return
	}
	if err := libs.SetCookie(c, "token", token.AccessToken, 3600, "/"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set cookie", "error": err.Error()})
		return
	}
	if err := libs.SetCookie(c, "refresh_token", token.RefreshToken, 3600, "/"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set cookie", "error": err.Error()})
		return
	}
	c.Header("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"error": false,
	})

}

//	func VerifyUserEndpoint(c *gin.Context) {
//		var email = c.Param("email")
//		var token = c.Param("token")
//
//		if email == "" || token == "" {
//			c.JSON(http.StatusBadRequest, gin.H{"message": "email and token are required", "error": true})
//			return
//		}
//
//		if err := services.VerifyUser(token, email); err != nil {
//			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error(), "error": true})
//			return
//		}
//
//		c.JSON(http.StatusOK, gin.H{"message": "User verified successfully", "error": false})
//	}
func CreatePasswordResetLinkEndpoint(c *gin.Context) {
	var PassWordReset struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&PassWordReset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "email is required",
		})
		return
	}

	resetToken, err := services.CreatePasswordReset(PassWordReset.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error(), "error": true})
		return
	}

	if resetToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to generate reset token", "error": true})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset email sent successfully",
		"error":   false,
	})

}

func ResetPasswordTokenVerifyEndpoint(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "token is required", "error": true})
		return
	}

	email, err := services.VerifyPasswordResetToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error(), "error": true})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token verified successfully",
		"error":   false,
		"email":   email,
	})

}

func ResetPasswordEndpoint(c *gin.Context) {
	type Request struct {
		Password string `json:"password" binding:"required"`
		Token    string `json:"token" binding:"required"`
	}

	var value Request

	if err := c.ShouldBindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error(), "error": true})
		return
	}
	fmt.Printf("Received token: %s and password: %s\n", value.Token, value.Password)
	if err := services.ResetPassword(value.Token, value.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to to reset password", "error": true})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
		"error":   false,
	})

}

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
		if err := libs.SetCookie(c, "admin_token", token.AccessToken, 3600, "/"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set cookie", "error": err.Error()})
			return
		}
		if err := libs.SetCookie(c, "refresh_token", token.RefreshToken, 3600, "/"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set cookie", "error": err.Error()})
			return
		}
		c.Header("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"error": false,
		})
		return
	}

	if err := libs.SetCookie(c, "token", token.AccessToken, 3600, "/"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set cookie", "error": err.Error()})
		return
	}
	if err := libs.SetCookie(c, "refresh_token", token.RefreshToken, 3600, "/"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set cookie", "error": err.Error()})
		return
	}
	c.Header("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"error": false,
	})
}

func LogoutUserEndpoint(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.SetCookie("admin_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "User logged out successfully", "error": false})
}

func RefreshTokenEndpoint(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Refresh token is required", "error": true})
		return
	}

	newTokens, err := services.RefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error(), "error": true})
		return
	}

	if newTokens.IsAdmin {
		if err := libs.SetCookie(c, "admin_token", newTokens.AccessToken, 3600, "/"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set cookie", "error": err.Error()})
			return
		}
		if err := libs.SetCookie(c, "refresh_token", newTokens.RefreshToken, 3600, "/"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set cookie", "error": err.Error()})
			return
		}
		c.Header("Authorization", fmt.Sprintf("Bearer %s", newTokens.AccessToken))
		c.JSON(http.StatusOK, gin.H{
			"token": newTokens,
			"error": false,
		})
		return
	}

	if err := libs.SetCookie(c, "token", newTokens.AccessToken, 3600, "/"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set cookie", "error": err.Error()})
		return
	}
	if err := libs.SetCookie(c, "refresh_token", newTokens.RefreshToken, 3600, "/"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to set cookie", "error": err.Error()})
		return
	}
	c.Header("Authorization", fmt.Sprintf("Bearer %s", newTokens.AccessToken))
	c.JSON(http.StatusOK, gin.H{
		"token": newTokens,
		"error": false,
	})
}
