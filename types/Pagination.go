package types

// PaginationRequest represents pagination parameters for requests
type PaginationRequest struct {
	Page  int `json:"page" form:"page" query:"page"`
	Limit int `json:"limit" form:"limit" query:"limit"`
}

// PaginationResponse represents pagination metadata for responses
type PaginationResponse struct {
	Page         int   `json:"page"`
	Limit        int   `json:"limit"`
	Total        int64 `json:"total"`
	TotalPages   int64 `json:"totalPages"`
	HasNextPage  bool  `json:"hasNextPage"`
	HasPrevPage  bool  `json:"hasPrevPage"`
	NextPage     *int  `json:"nextPage,omitempty"`
	PreviousPage *int  `json:"previousPage,omitempty"`
}

// NewPaginationResponse creates a new pagination response with calculated values
func NewPaginationResponse(page, limit int, total int64) *PaginationResponse {
	totalPages := (total + int64(limit) - 1) / int64(limit)

	pagination := &PaginationResponse{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNextPage: int64(page) < totalPages,
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

// ValidatePagination validates and normalizes pagination parameters
func ValidatePagination(page, limit int) (int, int) {
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 20 // default limit
	}

	if limit > 100 {
		limit = 100 // max limit
	}

	return page, limit
}

// GetOffset calculates the offset for database queries
func (p *PaginationRequest) GetOffset() int {
	page, _ := ValidatePagination(p.Page, p.Limit)
	return (page - 1) * p.Limit
}

// GetValidatedParams returns validated page and limit values
func (p *PaginationRequest) GetValidatedParams() (int, int) {
	return ValidatePagination(p.Page, p.Limit)
}
