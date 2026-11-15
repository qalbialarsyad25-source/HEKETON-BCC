package helper

import (
	apperrors "main/internal/shared/errors"
	"main/internal/shared/pagination"
)

// SuccessResponse is the body of every 2xx JSON response.
type SuccessResponse[T any] struct {
	Data       T                              `json:"data"`
	Pagination *pagination.PaginationResponse `json:"pagination,omitempty"`
}

// ErrorResponse is the body of every error response.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Issues  []apperrors.Issue `json:"issues,omitempty"`
}

// ErrorFrom is the body ErrorHandler writes for err.
func ErrorFrom(err *apperrors.AppError) ErrorResponse {
	return ErrorResponse{Error: ErrorBody{Code: err.Code, Message: err.Message, Issues: err.Issues}}
}
