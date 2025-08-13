package controllers

import (
	"net/http"

	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/gin-gonic/gin"
)

func FetchUserEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "user not found in context", "error": true})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	response, err := services.FetchUser(claims.UserID)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "user not found in context", "error": true})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":    response,
		"error":   false,
		"message": "User fetched successfully",
	})
}
