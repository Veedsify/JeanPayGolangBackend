package utils

import (
	"encoding/json"
	"net/http"

	"github.com/Veedsify/JeanPayGoBackend/constants"
)

// APIResponse represents a standard API response structure
type APIResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Pagination `json:"meta,omitempty"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page         int   `json:"page"`
	Limit        int   `json:"limit"`
	Total        int64 `json:"total"`
	TotalPages   int   `json:"totalPages"`
	HasNextPage  bool  `json:"hasNextPage"`
	HasPrevPage  bool  `json:"hasPrevPage"`
	NextPage     *int  `json:"nextPage,omitempty"`
	PreviousPage *int  `json:"previousPage,omitempty"`
}

// ErrorResponse represents an error response structure
type ErrorResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Code    string      `json:"code,omitempty"`
}

// NewPagination creates a new pagination metadata
func NewPagination(page, limit int, total int64) *Pagination {
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	pagination := &Pagination{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNextPage: page < totalPages,
		HasPrevPage: page > 1,
	}

	if pagination.HasNextPage {
		nextPage := page + 1
		pagination.NextPage = &nextPage
	}

	if pagination.HasPrevPage {
		prevPage := page - 1
		pagination.PreviousPage = &prevPage
	}

	return pagination
}

// WriteJSONResponse writes a JSON response to the http.ResponseWriter
func WriteJSONResponse(w http.ResponseWriter, statusCode int, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(response)
}

// WriteErrorResponse writes an error response
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string, details interface{}) error {
	response := ErrorResponse{
		Error:   true,
		Message: message,
		Details: details,
	}
	return WriteJSONResponse(w, statusCode, response)
}

// WriteSuccessResponse writes a success response
func WriteSuccessResponse(w http.ResponseWriter, message string, data interface{}) error {
	response := APIResponse{
		Error:   false,
		Message: message,
		Data:    data,
	}
	return WriteJSONResponse(w, constants.StatusOK, response)
}

// WriteCreatedResponse writes a created response
func WriteCreatedResponse(w http.ResponseWriter, message string, data interface{}) error {
	response := APIResponse{
		Error:   false,
		Message: message,
		Data:    data,
	}
	return WriteJSONResponse(w, constants.StatusCreated, response)
}

// WritePaginatedResponse writes a paginated response
func WritePaginatedResponse(w http.ResponseWriter, message string, data interface{}, pagination *Pagination) error {
	response := PaginatedResponse{
		Error:   false,
		Message: message,
		Data:    data,
		Meta:    pagination,
	}
	return WriteJSONResponse(w, constants.StatusOK, response)
}

// WriteBadRequestResponse writes a bad request error response
func WriteBadRequestResponse(w http.ResponseWriter, message string, details interface{}) error {
	return WriteErrorResponse(w, constants.StatusBadRequest, message, details)
}

// WriteUnauthorizedResponse writes an unauthorized error response
func WriteUnauthorizedResponse(w http.ResponseWriter, message string) error {
	if message == "" {
		message = "Unauthorized access"
	}
	return WriteErrorResponse(w, constants.StatusUnauthorized, message, nil)
}

// WriteForbiddenResponse writes a forbidden error response
func WriteForbiddenResponse(w http.ResponseWriter, message string) error {
	if message == "" {
		message = "Access forbidden"
	}
	return WriteErrorResponse(w, constants.StatusForbidden, message, nil)
}

// WriteNotFoundResponse writes a not found error response
func WriteNotFoundResponse(w http.ResponseWriter, message string) error {
	if message == "" {
		message = "Resource not found"
	}
	return WriteErrorResponse(w, constants.StatusNotFound, message, nil)
}

// WriteConflictResponse writes a conflict error response
func WriteConflictResponse(w http.ResponseWriter, message string, details interface{}) error {
	return WriteErrorResponse(w, constants.StatusConflict, message, details)
}

// WriteUnprocessableEntityResponse writes an unprocessable entity error response
func WriteUnprocessableEntityResponse(w http.ResponseWriter, message string, details interface{}) error {
	return WriteErrorResponse(w, constants.StatusUnprocessableEntity, message, details)
}

// WriteTooManyRequestsResponse writes a too many requests error response
func WriteTooManyRequestsResponse(w http.ResponseWriter, message string) error {
	if message == "" {
		message = "Rate limit exceeded"
	}
	return WriteErrorResponse(w, constants.StatusTooManyRequests, message, nil)
}

// WriteInternalServerErrorResponse writes an internal server error response
func WriteInternalServerErrorResponse(w http.ResponseWriter, message string, details interface{}) error {
	if message == "" {
		message = "Internal server error"
	}
	return WriteErrorResponse(w, constants.StatusInternalServerError, message, details)
}

// WriteServiceUnavailableResponse writes a service unavailable error response
func WriteServiceUnavailableResponse(w http.ResponseWriter, message string) error {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	return WriteErrorResponse(w, constants.StatusServiceUnavailable, message, nil)
}

// WriteValidationErrorResponse writes a validation error response
func WriteValidationErrorResponse(w http.ResponseWriter, errors ValidationErrors) error {
	return WriteUnprocessableEntityResponse(w, "Validation failed", errors)
}

// WriteNoContentResponse writes a no content response
func WriteNoContentResponse(w http.ResponseWriter) error {
	w.WriteHeader(constants.StatusNoContent)
	return nil
}

// WriteAcceptedResponse writes an accepted response
func WriteAcceptedResponse(w http.ResponseWriter, message string, data interface{}) error {
	response := APIResponse{
		Error:   false,
		Message: message,
		Data:    data,
	}
	return WriteJSONResponse(w, constants.StatusAccepted, response)
}

// ParseJSONBody parses JSON request body
func ParseJSONBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// GetQueryParam gets a query parameter value
func GetQueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// GetQueryParamWithDefault gets a query parameter value with a default
func GetQueryParamWithDefault(r *http.Request, key, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// SetCORSHeaders sets CORS headers
func SetCORSHeaders(w http.ResponseWriter, allowedOrigins []string) {
	origin := "*"
	if len(allowedOrigins) > 0 && allowedOrigins[0] != "*" {
		origin = allowedOrigins[0]
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

// HandlePreflight handles preflight OPTIONS requests
func HandlePreflight(w http.ResponseWriter, allowedOrigins []string) {
	SetCORSHeaders(w, allowedOrigins)
	w.WriteHeader(constants.StatusOK)
}
