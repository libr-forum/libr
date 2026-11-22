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
		log.Fatalf("Failed to open log file: %v", err)
	}
}

func LogToFile(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.SetOutput(logFile)
	log.Println("[" + timestamp + "] " + message)
}
