package pagination

import "math"

//nolint:revive // the name PaginationResponse is kept from the original codebase on purpose (see docs/STYLE_CHANGES.md)
type PaginationResponse struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	HasNext    bool `json:"hasNext"`
	HasPrev    bool `json:"hasPrev"`
	TotalItems int  `json:"totalItems"`
	TotalPages int  `json:"totalPages"`
}

// Pagination is the page requested by the client, filled by middleware.HandlePagination.
type Pagination struct {
	Offset int `form:"-"`
	Page   int `form:"page"`
	Limit  int `form:"limit"`
}

// GetPaginationResponse builds the pagination metadata for a list of totalItems items.
func GetPaginationResponse(pagination Pagination, totalItems int) *PaginationResponse {
	totalPages := 0
	if pagination.Limit > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(pagination.Limit)))
	}

	return &PaginationResponse{
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1 && totalPages > 0,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
