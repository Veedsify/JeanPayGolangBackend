package services

import (
	"fmt"
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database"
	"github.com/Veedsify/JeanPayGoBackend/database/models"
	"github.com/Veedsify/JeanPayGoBackend/jobs"
	"github.com/Veedsify/JeanPayGoBackend/types"
)

func SendContactMessage(contact types.ContactRequest) (bool, error) {

	newContact := &models.Contact{
		Email:       contact.Email,
		FullName:    contact.FullName,
		PhoneNumber: contact.PhoneNumber,
		Category:    contact.Category,
		Subject:     contact.Subject,
		Message:     contact.Message,
		Delivered:   false,
	}

	if err := database.DB.Create(newContact).Error; err != nil {
		return true, fmt.Errorf("failed to save contact info: %v", err)
	}

	var user models.User

	if err := database.DB.Where("is_admin = ?", true).First(&user).Error; err != nil {
		return true, fmt.Errorf("failed to get admin  account: %v", err)
	}
	emailData := jobs.EmailJobPayload{
		To:         []string{user.Email},
		Subject:    contact.Subject,
		TemplateID: "general_email",
		Data: map[string]any{
			"ContactEmail": &contact.Email,
			"Subject":      &contact.Subject,
			"Message":      &contact.Message,
			"Category":     &contact.Category,
			"FullName":     &contact.FullName,
			"Phone":        &contact.PhoneNumber,
		},
		Priority: "high",
	}

	emailJob := jobs.NewEmailJobClient()
	err := emailJob.EnqueueScheduledEmail(emailData, time.Now().Add(time.Second))

	if err != nil {
		return true, fmt.Errorf("failed to send email to admin  account: %v", err)
	}

	return false, nil
}
