package middlewares

import (
	"net/http"

	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/gin-gonic/gin"
)

func CheckUserIsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   true,
				"message": "User not authenticated",
			})
			return
		}
		IsAdmin := claims.(*libs.JWTClaims).IsAdmin
		if !IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   true,
				"message": "You do not have permission to access this resource",
			})
			return
		}
		c.Next()
	}
}
