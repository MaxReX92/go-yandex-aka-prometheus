package logger

import (
	"fmt"
	"log"
)

// Info log information message.
func Info(message string) {
	log.Printf("[INFO]: %v\r\n", message)
}

// InfoFormat log information message with custom format arguments.
func InfoFormat(format string, v ...any) {
	Info(fmt.Sprintf(format, v...))
}

// Warn log warning message.
func Warn(message string) {
	log.Printf("[WARN]: %v\r\n", message)
}

// WarnFormat log warning message with custom format arguments.
func WarnFormat(format string, v ...any) {
	Warn(fmt.Sprintf(format, v...))
}

// Error log error message.
func Error(message string) {
	log.Printf("[ERROR]: %v\r\n", message)
}

// ErrorObj log error object.
func ErrorObj(err error) {
	ErrorFormat("%v", err)
}

// ErrorFormat log error message with custom format arguments.
func ErrorFormat(format string, v ...any) {
	Error(fmt.Sprintf(format, v...))
}

// WrapError log error and return wrapped error object.
func WrapError(message string, err error) error {
	wrap := fmt.Errorf("failed to "+message+": %w", err) //nolint:goerr113
	ErrorObj(wrap)

	return wrap
}
