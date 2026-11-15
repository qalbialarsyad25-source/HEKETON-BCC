package pagination

import (
	"math"
	"strconv"
)

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

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

type Params struct {
	Page  int
	Limit int
}

func NewParams(page, limit int) Params {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return Params{Page: page, Limit: limit}
}

// ParamsFromQuery reads the raw ?page= and ?limit= values. A missing or invalid value
// falls back to the default instead of failing the request.
func ParamsFromQuery(page, limit string) Params {
	p, _ := strconv.Atoi(page)
	l, _ := strconv.Atoi(limit)

	return NewParams(p, l)
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.Limit
}

type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

func NewMeta(params Params, totalItems int64) Meta {
	totalPages := int((totalItems + int64(params.Limit) - 1) / int64(params.Limit))

	return Meta{
		Page:       params.Page,
		Limit:      params.Limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

type Result[T any] struct {
	Items []T
	Meta  Meta
}

func NewResult[T any](items []T, params Params, totalItems int64) *Result[T] {
	return &Result[T]{Items: items, Meta: NewMeta(params, totalItems)}
}
