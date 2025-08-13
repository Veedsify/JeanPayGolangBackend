package middlewares

import (
	"net/http"

	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/gin-gonic/gin"
)

type Error struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}

func AuthMiddleware(jwtService *libs.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := jwtService.ExtractTokenFromHeader(c.GetHeader("Authorization"))
		if err != nil {
			tokenString, err = c.Cookie("token") // or whatever cookie name you use
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "no token provided"})
				c.Abort()
				return
			}
		}

		claims, err := jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user", claims)
		c.Next()
	}
}
