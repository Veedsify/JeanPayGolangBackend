package types

import (
	"time"

	"github.com/Veedsify/JeanPayGoBackend/database/models"
)

type CreateAdminLogRequest struct {
	AdminID   string                 `json:"admin_id" form:"admin_id" binding:"required"`
	Action    string                 `json:"action" form:"action" binding:"required"`
	Target    string                 `json:"target" form:"target" binding:"required"`
	TargetID  string                 `json:"target_id" form:"target_id" binding:"required"`
	Details   map[string]interface{} `json:"details" form:"details"`
	IPAddress string                 `json:"ip_address" form:"ip_address"`
	UserAgent string                 `json:"user_agent" form:"user_agent"`
}

type AdminLogResponse struct {
	ID        uint32    `json:"id"`
	AdminID   uint32    `json:"admin_id"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	TargetID  string    `json:"target_id"`
	Details   string    `json:"details"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetAdminLogsRequest struct {
	AdminID string `form:"admin_id"`
	Action  string `form:"action"`
	Target  string `form:"target"`
	Page    int    `form:"page"`
	Limit   int    `form:"limit"`
}

type GetAdminLogsResponse struct {
	Logs       []AdminLogResponse `json:"logs"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}

func ToAdminLogResponse(adminLog *models.AdminLog) AdminLogResponse {
	return AdminLogResponse{
		ID:        uint32(adminLog.ID),
		AdminID:   adminLog.AdminID,
		Action:    adminLog.Action,
		Target:    adminLog.Target,
		TargetID:  adminLog.TargetID,
		Details:   adminLog.Details,
		IPAddress: adminLog.IPAddress,
		UserAgent: adminLog.UserAgent,
		CreatedAt: adminLog.CreatedAt,
		UpdatedAt: adminLog.UpdatedAt,
	}
}

func ToAdminLogsResponse(adminLogs []models.AdminLog) []AdminLogResponse {
	var response []AdminLogResponse
	for _, adminLog := range adminLogs {
		response = append(response, ToAdminLogResponse(&adminLog))
	}
	return response
}
