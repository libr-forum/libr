package core

import (
	"encoding/json"
	"libr/keycache"
	"libr/types"
	"log"
	"sort"

	"github.com/Arnav-Agrawal-987/crypto/cryptoutils"
)

func CreateMsgCert(message string, ts int64, modcertList []types.ModCert) types.MsgCert {
	_, privKey := keycache.PubKey, keycache.PrivKey

	sort.SliceStable(modcertList, func(i, j int) bool {
		return modcertList[i].PublicKey < modcertList[j].PublicKey
	})

	dataToSign := types.DataToSign{
		Content:   message,
		Timestamp: ts,
		ModCerts:  modcertList, // sorted before signing
	}

	jsonBytes, _ := json.Marshal(dataToSign)
	pubKeyStr, sign, err := cryptoutils.SignMessage(privKey, string(jsonBytes))
	if err != nil {
		log.Fatalf("failed to sign message: %v", err)
	}

	msgCert := types.MsgCert{
		PublicKey: pubKeyStr,
		Msg: types.Msg{
			Content: message,
			Ts:      ts,
		},
		ModCerts: modcertList,
		Sign:     sign,
	}

	return msgCert

}
