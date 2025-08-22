package main

import (
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/initializers"
	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func init() {
	go initializers.InitializeQueueServer()
	database.InitDB()
}

func main() {
	allowedDomains := []string{
		libs.GetEnvOrDefault("FRONTEND_URL", "http://localhost:3000"),
		libs.GetEnvOrDefault("ADMIN_URL", "http://localhost:3001"),
		"http://app.test:3000",
		"http://app.test:3001",
	}
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedDomains,
		AllowMethods:     []string{"PUT", "PATCH", "POST", "DELETE", "OPTIONS", "GET"},
		AllowHeaders:     []string{"Origin", "Cookie", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	routes.ApiRoutes(router)
	// Print all routes before running
	router.Run() // listen and serve on 0.0.0.0:8080
}
