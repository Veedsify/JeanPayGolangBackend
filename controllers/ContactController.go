package controllers

import (
	"fmt"
	"net/http"

	"github.com/Veedsify/JeanPayGoBackend/services"
	"github.com/Veedsify/JeanPayGoBackend/types"
	"github.com/gin-gonic/gin"
)

func ContactInformation(c *gin.Context) {
	var contact types.ContactRequest

	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": fmt.Sprintf("All Fields Are Required %v", err),
		})
		return
	}

	_, err := services.SendContactMessage(contact)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Failed To Send Contact Message At This Time",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Your Request Has Been Submitted",
	})
}
