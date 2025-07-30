package core

import (
	"encoding/json"
	"log"
	"sort"

	"github.com/devlup-labs/Libr/core/mod_client/keycache"
	"github.com/devlup-labs/Libr/core/mod_client/types"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
)

func CreateRepCert(originalsender string, message string, ts int64, modcertList []types.ModCert) types.ReportCert {
	_, privKey := keycache.PubKey, keycache.PrivKey

	sort.SliceStable(modcertList, func(i, j int) bool {
		return modcertList[i].PublicKey < modcertList[j].PublicKey
	})

	dataToSign := map[string]interface{}{
		"PublicKey": originalsender,
		"Content":   message,
		"Ts":        ts,
		"ModCerts":  modcertList,
	}

	jsonBytes, _ := json.Marshal(dataToSign)
	pubKeyStr, sign, err := cryptoutils.SignMessage(privKey, string(jsonBytes))
	if err != nil {
		log.Fatalf("failed to sign message: %v", err)
	}

	repCert := types.ReportCert{
		PublicKey: pubKeyStr,
		ReportMsg: types.ReportMsg{
			Msg: types.Msg{
				Content: message,
				Ts:      ts,
			},
			PublicKey: originalsender,
		},
		ModCerts: modcertList,
		Sign:     sign,
	}

	return repCert

}
