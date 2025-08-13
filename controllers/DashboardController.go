package controllers

import (
	"net/http"

	"github.com/Veedsify/JeanPayGoBackend/libs"
	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/gin-gonic/gin"
)

// GetDashboardOverviewEndpoint retrieves dashboard overview for authenticated user
func GetDashboardOverviewEndpoint(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}
	userID := claims.(*libs.JWTClaims).ID

	overview, err := services.GetDashboardOverview(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Dashboard overview retrieved successfully",
		"data":    overview,
	})
}

// GetDashboardStatsEndpoint retrieves detailed dashboard statistics for authenticated user
func GetDashboardStatsEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)
	stats, err := services.GetDashboardStats(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Dashboard statistics retrieved successfully",
		"data":    stats,
	})
}

// GetUserTransactionSummaryEndpoint retrieves transaction summary for authenticated user
func GetUserTransactionSummaryEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)
	summary, err := services.GetUserTransactionSummary(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Transaction summary retrieved successfully",
		"data":    summary,
	})
}

// GetTransactionStatsEndpoint retrieves transaction statistics
func GetTransactionStatsEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)
	period := c.DefaultQuery("period", "month")

	stats, err := services.GetDashboardTransactionStats(claims.UserID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Transaction statistics retrieved successfully",
		"data":    stats,
	})
}

// GetDashboardChartsDataEndpoint retrieves chart data for dashboard
func GetDashboardChartsDataEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)
	stats, err := services.GetDashboardStats(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Dashboard chart data retrieved successfully",
		"data":    stats.ChartData,
	})
}

// GetDashboardSummaryEndpoint retrieves dashboard summary
func GetDashboardSummaryEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	summary, err := services.GetDashboardSummary(claims.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Dashboard summary retrieved successfully",
		"data":    summary,
	})
}

// GetRecentActivityEndpoint retrieves recent activity
func GetRecentActivityEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	activity, err := services.GetRecentActivity(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Recent activity retrieved successfully",
		"data":    activity,
	})
}

// GetWalletOverviewEndpoint retrieves wallet overview
func GetWalletOverviewEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	overview, err := services.GetWalletOverview(claims.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Wallet overview retrieved successfully",
		"data":    overview,
	})
}

// GetConversionStatsEndpoint retrieves conversion statistics
func GetConversionStatsEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	stats, err := services.GetConversionStats(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Conversion statistics retrieved successfully",
		"data":    stats,
	})
}

// GetMonthlyStatsEndpoint retrieves monthly statistics
func GetMonthlyStatsEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	stats, err := services.GetMonthlyStats(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Monthly statistics retrieved successfully",
		"data":    stats,
	})
}

// GetTransactionTrendsEndpoint retrieves transaction trends
func GetTransactionTrendsEndpoint(c *gin.Context) {
	claimsAny, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "User not authenticated",
		})
		return
	}

	claims := claimsAny.(*libs.JWTClaims)

	trends, err := services.GetTransactionTrends(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Transaction trends retrieved successfully",
		"data":    trends,
	})
}
