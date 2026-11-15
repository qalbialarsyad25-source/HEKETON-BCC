package middleware

import (
	"math"

	"main/internal/shared/pagination"

	"github.com/gin-gonic/gin"
)

// PaginationKey is the gin context key holding the pagination.Pagination set by HandlePagination.
const PaginationKey = "pagination"

// HandlePagination reads ?page= and ?limit= into a pagination.Pagination under PaginationKey.
// It never fails: an unreadable query or a value ≤ 0 falls back to page 1 / defaultLimit,
// and a limit above maxLimit is clamped to maxLimit (a huge page is clamped too, so the offset can't overflow).
func HandlePagination(defaultLimit int, maxLimit int) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var paging pagination.Pagination

		if err := ctx.ShouldBindQuery(&paging); err != nil {
			paging.Limit = defaultLimit
			paging.Page = 1
		}

		if paging.Limit <= 0 {
			paging.Limit = defaultLimit
		}

		if paging.Limit > maxLimit {
			paging.Limit = maxLimit
		}

		if paging.Page <= 0 {
			paging.Page = 1
		}

		// Keep (Page-1)*Limit from overflowing into a negative offset.
		if maxPage := math.MaxInt / paging.Limit; paging.Page > maxPage {
			paging.Page = maxPage
		}

		paging.Offset = (paging.Page - 1) * paging.Limit

		ctx.Set(PaginationKey, paging)
		ctx.Next()
	}
}
