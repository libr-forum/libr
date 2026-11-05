package logger

import (
	"log"
)

func LogToFile(message string) {
	// For mobile, we log to stdout/stderr which gomobile redirects to logcat.
	log.Printf("[LIBR-GO] %s", message)
}
