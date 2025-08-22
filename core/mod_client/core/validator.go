package core

import (
	"encoding/json"
	"strings"

	"github.com/libr-forum/Libr/core/mod_client/types"
)

func IsValidMessage(content string) bool {
	if len(content) == 0 || len(content) > 500 {
		return false
	}
	if strings.ContainsAny(content, "<>{}") {
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
