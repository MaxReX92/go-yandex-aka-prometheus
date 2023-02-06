package logger

import (
	"fmt"
	"log"
)

func Info(message string) {
	log.Println(fmt.Sprintf("[INFO]: %v", message))
}

func InfoFormat(format string, v ...any) {
	Info(fmt.Sprintf(format, v...))
}

func Warn(message string) {
	log.Println(fmt.Sprintf("[WARN]: %v", message))
}

func WarnFormat(format string, v ...any) {
	Warn(fmt.Sprintf(format, v...))
}

func Error(message string) {
	log.Println(fmt.Sprintf("[ERROR]: %v", message))
}

func ErrorFormat(format string, v ...any) {
	Error(fmt.Sprintf(format, v...))
}
