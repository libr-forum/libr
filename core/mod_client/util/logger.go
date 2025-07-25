package util

import (
	"encoding/json"
	"sort"

	"github.com/devlup-labs/Libr/core/mod_client/types"
)

func CanonicalizeMsg(msg types.Msg) (string, error) {
	canonical, err := json.Marshal(struct {
		Content string `json:"content"`
		Ts      int64  `json:"ts"`
	}{
		Content: msg.Content,
		Ts:      msg.Ts,
	})
	if err != nil {
		return "", err
	}

	return string(canonical), nil
}

func CanonicalizeMsgCert(msgcert types.MsgCert) (string, error) {

	sort.SliceStable(msgcert.ModCerts, func(i, j int) bool {
		return msgcert.ModCerts[i].PublicKey < msgcert.ModCerts[j].PublicKey
	})

	// Step 3: Marshal to JSON
	canonicalBytes, err := json.Marshal(msgcert)
	if err != nil {
		return "", err
	}

	return string(canonicalBytes), nil
}
