package middleware

import (
	"errors"

	apperrors "main/internal/shared/errors"
	"main/internal/shared/ports/logger"

	"github.com/gin-gonic/gin"
)

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Issues  []apperrors.Issue `json:"issues,omitempty"`
}

// ErrorHandler renders the last error a handler or middleware attached with c.Error as
// {"error": {"code", "message", "issues"}}. An *AppError keeps its status and code; anything
// else becomes a generic 500 and its cause is only logged. It does nothing when the handler
// already wrote a response (or upgraded the connection). Register it before the routes.
func ErrorHandler(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		err := c.Errors.Last().Err

		var appErr *apperrors.AppError
		if !errors.As(err, &appErr) {
			appErr = apperrors.Internal(err)
		}

		if log != nil {
			log.ErrorLog(c.GetHeader("X-Request-ID"), c.Request.Method, c.Request.URL.Path, appErr.StatusCode, appErr.Code, err)
		}

		c.AbortWithStatusJSON(appErr.StatusCode, gin.H{
			"error": errorBody{
				Code:    appErr.Code,
				Message: appErr.Message,
				Issues:  appErr.Issues,
			},
		})
	}
}
