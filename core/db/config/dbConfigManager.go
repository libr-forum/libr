package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/libr-forum/Libr/core/crypto/logger"
	"github.com/libr-forum/Libr/core/db/internal/models"
)

func GetDBConfigPath() string {
	var path string

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		path = filepath.Join(appData, "libr", "dbconfig", "dbconfig.json")
	case "darwin":
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, "Library", "Application Support", "libr", "dbconfig", "dbconfig.json")
	case "linux":
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, ".config", "libr", "dbconfig", "dbconfig.json")
	default:
		path = filepath.Join("dbconfig", "dbconfig.json")
	}

	return path
}


func ReadDBConfigFile() (models.DBConfig, error) {
	path := GetDBConfigPath()
	fmt.Println(path)
	   file, err := os.Open(path)
	   if err != nil {
		   logger.LogToFile("failed to open dbconfig.json")
		   return models.DBConfig{}, fmt.Errorf("failed to open dbconfig.json: %w", err)
	   }
	   defer file.Close()

	   // Debug: print raw file contents
	   fileInfo, _ := file.Stat()
	   fileSize := fileInfo.Size()
	   rawBytes := make([]byte, fileSize)
	   _, err = file.ReadAt(rawBytes, 0)
	   if err != nil {
		   fmt.Println("[DEBUG] Error reading raw file bytes:", err)
	   } else {
		   fmt.Println("[DEBUG] Raw dbconfig.json contents:")
		   fmt.Println(string(rawBytes))
	   }

	   // Reset file pointer to beginning for decoding
	   file.Seek(0, 0)

	   var configObj models.DBConfig
	   if err := json.NewDecoder(file).Decode(&configObj); err != nil {
		   logger.LogToFile("[DEBUG]Failed to decode dbconfig.json")
		   fmt.Println("[DEBUG] JSON decode error:", err)
		   return models.DBConfig{}, fmt.Errorf("failed to decode dbconfig.json: %w", err)
	   }

	   fmt.Println("[DEBUG] Decoded configObj:", configObj)

	   // Set DBtype based on API_KEY
	   if configObj.API_KEY == "" {
		   DBtype = "normal"
	   } else {
		   DBtype = "boot"
	   }

	   return configObj, nil
}