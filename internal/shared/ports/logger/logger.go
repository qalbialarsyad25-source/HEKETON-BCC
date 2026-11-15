package ports

import "time"

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	// RequestLog writes one line per finished request.
	RequestLog(requestID string, method string, path string, statusCode int, start time.Time)
	// ErrorLog writes one line per failed request: a warning for 4xx, an error for 5xx.
	ErrorLog(requestID string, method string, path string, statusCode int, code string, err error)
}
