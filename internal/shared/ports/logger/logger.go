package logger

import "time"

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	RequestLog(requestID string, method string, path string, statusCode int, start time.Time)
	ErrorLog(requestID string, method string, path string, statusCode int, code string, err error)
}
