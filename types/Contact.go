package types

type ContactRequest struct {
	Email       string `json:"email" binding:"required,email"`
	FullName    string `json:"full_name" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"omitempty"`
	Category    string `json:"category" binding:"required"`
	Message     string `json:"message" binding:"required"`
	Subject     string `json:"subject" binding:"required"`
}
