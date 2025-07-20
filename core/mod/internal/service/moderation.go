package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/devlup-labs/Libr/core/mod/models"
	"github.com/joho/godotenv"
)

var forbidden = []string{
	"bitch", "asshole", "arsehole", "fuck", "chutiya", "bhenchod", "nigga", 
}

type ModelFunc func(content string) (bool, error)

func init() {
	_ = godotenv.Load() // Load .env file if present

}

func ModerateMsg(msg models.UserMsg) (string, error) {

	for _, word := range forbidden {
		if msg.Content == word {
			return "0", nil
		}
	}

	clean, err := AnalyzeWithGoogleNLP(msg.Content)
	if err != nil {
		return "", err
	}
	if clean {
		return "1", nil
	}
	return "0", nil
}

func SelectModel(method string) (ModelFunc, error) {
	return AnalyzeWithGoogleNLP, nil
}

func AnalyzeContent(content string, fn ModelFunc) (string, error) {
	clean, err := fn(content)
	if err != nil {
		return "error", err
	}
	if clean {
		return "1", nil
	}
	return "0", nil
}

// Category holds moderation category info
type Category struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
}

// ParseThresholds reads thresholds only from the environment
func ParseThresholds() map[string]float64 {
	thresholds := make(map[string]float64)

	envVal := os.Getenv("THRESHOLDS")
	if envVal == "" {
		fmt.Println("No thresholds found in environment")
		return thresholds // return empty if not set
	}

	fmt.Println("Loaded THRESHOLDS from env:", envVal)

	pairs := strings.Split(envVal, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, ":")
		if len(kv) == 2 {
			conf, err := strconv.ParseFloat(kv[1], 64)
			if err == nil {
				thresholds[kv[0]] = conf
			} else {
				fmt.Printf("Invalid threshold value for %s: %s\n", kv[0], kv[1])
			}
		}
	}
	return thresholds
}

// AnalyzeWithGoogleNLP uses the ModerateText API
func AnalyzeWithGoogleNLP(content string) (bool, error) {
	apiKey := os.Getenv("GOOGLE_NLP_API_KEY")
	if apiKey == "" {
		return false, fmt.Errorf("missing GOOGLE_NLP_API_KEY in environment")
	}

	// Prepare request
	payload := map[string]interface{}{
		"document": map[string]interface{}{
			"type":    "PLAIN_TEXT",
			"content": content,
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	url := fmt.Sprintf("https://language.googleapis.com/v1/documents:moderateText?key=%s", apiKey)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Google NLP API error: %s", string(body))
	}

	var result struct {
		ModerationCategories []struct {
			Name       string  `json:"name"`
			Confidence float64 `json:"confidence"`
		} `json:"moderationCategories"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return false, err
	}

	thresholds := ParseThresholds()

	fmt.Println(result)

	for _, cat := range result.ModerationCategories {
		if th, exists := thresholds[cat.Name]; exists {
			if cat.Confidence >= th {
				fmt.Printf("Blocked category: %s (%.2f ≥ %.2f)\n", cat.Name, cat.Confidence, th)
				return false, nil
			}
		}
	}
	return true, nil
}
