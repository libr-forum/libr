package util

import (
	"bytes"
	"encoding/json"
	"libr/types"
	"strings"
)

func CanonicalizeMsg(msg types.Msg) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "") // compact output

	err := enc.Encode(struct {
		Content string `json:"content"`
		Ts      int64  `json:"ts"`
	}{
		Content: msg.Content,
		Ts:      msg.Ts,
	})
	if err != nil {
		return "", err
	}

	// Remove newline at the end if any
	return strings.TrimSpace(buf.String()), nil
}
