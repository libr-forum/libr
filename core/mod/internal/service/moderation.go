package service

import (
	"fmt"
	"strings"

	"github.com/devlup-labs/Libr/core/mod/models"
)

var forbidden = []string{
	"bitch", "asshole", "arsehole", "fuck", "chutiya", "bhenchod", "nigga",
}

type ModelFunc func(content string) (bool, error)

func ModerateMsg(msg models.UserMsg) (string, error) {
	clean, err := AnalyzeWithKeywordFilter(msg.Content)
	if err != nil {
		return "", err
	}
	if clean {
		fmt.Println("Message:", msg.Content, "\nTimestamp:", msg.TimeStamp, "\nStatus: Approved")
		return "1", nil
	}
	fmt.Println("Message:", msg.Content, "\nTimestamp:", msg.TimeStamp, "\nStatus: Rejected")
	return "0", nil
}

func AnalyzeWithKeywordFilter(content string) (bool, error) {
	lc := strings.ToLower(content)
	for _, bad := range forbidden {
		if strings.Contains(lc, bad) {
			return false, nil
		}
	}
	return true, nil
}

func SelectModel(method string) (ModelFunc, error) {
	// Only one method for now.
	return AnalyzeWithKeywordFilter, nil
}

func AnalyzeContent(content string, fn ModelFunc) (string, error) {
	clean, err := fn(content)
	if err != nil {
		return "", err
	}
	if clean {
		return "1", nil
	}
	return "0", nil
}
