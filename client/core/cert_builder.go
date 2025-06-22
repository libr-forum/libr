package core

import (
	"bytes"
	"encoding/json"
	"libr/types"
	"log"
	"sort"

	"github.com/Arnav-Agrawal-987/crypto/cryptoutils"
)

func CreateMsgCert(message string, ts int64, modcertList []types.ModCert) types.MsgCert {
	pubKey, privKey, err := cryptoutils.LoadKeys() //crypto
	if err != nil {
		log.Fatalf("failed to load keys: %v", err)
	}

	sort.SliceStable(modcertList, func(i, j int) bool {
		return bytes.Compare(modcertList[i].PublicKey, modcertList[j].PublicKey) < 0
	})

	dataToSign := types.DataToSign{
		Content:   message,
		Timestamp: ts,
		ModCerts:  modcertList, // sorted before signing
	}
	jsonBytes, _ := json.Marshal(dataToSign)
	sign, err := cryptoutils.SignMessage(privKey, string(jsonBytes))
	if err != nil {
		log.Fatalf("failed to sign message: %v", err)
	}

	msgCert := types.MsgCert{
		PublicKey: pubKey,
		Msg: types.Msg{
			Content: message,
			Ts:      ts,
		},
		ModCerts: modcertList,
		Sign:     sign,
	}

	return msgCert

}
