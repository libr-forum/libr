package logger

import (
	"log"
	"os"
	"time"
)

var logFile *os.File

func init() {
	var err error
	logFile, err = os.OpenFile("DEBUG_LOGS.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Not fatal: on mobile the working directory is not writable.
		log.Printf("debug log file unavailable, using stderr: %v", err)
		logFile = nil
	}
}

func LogToFile(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if logFile != nil {
		log.SetOutput(logFile)
	}
	log.Println("[" + timestamp + "] " + message)
}
