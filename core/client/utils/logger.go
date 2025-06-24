package util

import (
	"encoding/json"
	"sort"

	"github.com/devlup-labs/Libr/core/client/types"
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

func CanonicalizeMsgCert(msg types.Msg, modCerts []types.ModCert) (string, error) {
	// Step 1: Sort the ModCerts by PublicKey (to ensure deterministic signing)
	sort.SliceStable(modCerts, func(i, j int) bool {
		return modCerts[i].PublicKey < modCerts[j].PublicKey
	})

	// Step 2: Build a canonical struct
	canonical := struct {
		Msg      types.Msg       `json:"msg"`
		ModCerts []types.ModCert `json:"modCerts"`
	}{
		Msg:      msg,
		ModCerts: modCerts,
	}

	// Step 3: Marshal to JSON
	canonicalBytes, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}

	return string(canonicalBytes), nil
}
