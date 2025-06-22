package core

import (
	"strings"
)

func IsValidMessage(content string) bool {
	trimmedContent := strings.TrimSpace(content)
	if len(trimmedContent) == 0 || len(trimmedContent) > 500 {
		return false
	}
	if strings.ContainsAny(trimmedContent, "<>{}") {
		return false
	}
	return true
}
