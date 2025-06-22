package core

import (
	"encoding/json"
	"libr/types"
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

func CanonicalizeMsg(msg types.Msg) (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
