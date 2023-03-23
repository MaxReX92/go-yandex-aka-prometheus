package logger

import (
	"fmt"
	"log"
)

func Info(message string) {
	log.Printf("[INFO]: %v\r\n", message)
}

func InfoFormat(format string, v ...any) {
	Info(fmt.Sprintf(format, v...))
}

func Warn(message string) {
	log.Printf("[WARN]: %v\r\n", message)
}

func WarnFormat(format string, v ...any) {
	Warn(fmt.Sprintf(format, v...))
}

func Error(message string) {
	log.Printf("[ERROR]: %v\r\n", message)
}

func ErrorObj(err error) {
	ErrorFormat("%v", err)
}

func ErrorFormat(format string, v ...any) {
	Error(fmt.Sprintf(format, v...))
}

func WrapError(message string, err error) error {
	wrap := fmt.Errorf("failed to "+message+": %w", err) //nolint:goerr113
	ErrorObj(wrap)

	return wrap
}
