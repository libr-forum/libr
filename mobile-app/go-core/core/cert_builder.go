package core

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"log"
	"sort"

	"github.com/libr-forum/Libr/core/mod_client/keycache"
	"github.com/libr-forum/Libr/core/mod_client/types"

	"github.com/libr-forum/Libr/core/crypto/cryptoutils"
)

func CreateMsgCert(message string, ts int64, modcertList []types.ModCert) types.MsgCert {
	return CreateMsgCertWithPrivateKey(message, ts, modcertList, keycache.PrivKey)
}

func CreateMsgCertWithPrivateKey(message string, ts int64, modcertList []types.ModCert, privKey ed25519.PrivateKey) types.MsgCert {

	sort.SliceStable(modcertList, func(i, j int) bool {
		return modcertList[i].PublicKey < modcertList[j].PublicKey
	})

	dataToSign := types.DataToSign{
		Content:   message,
		Timestamp: ts,
		ModCerts:  modcertList,
	}

	jsonBytes, _ := json.Marshal(dataToSign)
	pubKeyStr, sign, err := cryptoutils.SignMessage(privKey, string(jsonBytes))
	fmt.Printf("[CreateMsgCert] public=%s sign=%s\n", pubKeyStr, sign)
	if err != nil {
		log.Printf("failed to sign message: %v", err)
	}

	return types.MsgCert{
		PublicKey: pubKeyStr,
		Msg: types.Msg{
			Content: message,
			Ts:      ts,
		},
		ModCerts: modcertList,
		Sign:     sign,
	}
}

func CreateRepCert(msgcert types.MsgCert, modcertList []types.ModCert, mode string) types.ReportCert {
	sort.SliceStable(modcertList, func(i, j int) bool {
		return modcertList[i].PublicKey < modcertList[j].PublicKey
	})

	return types.ReportCert{
		Msgcert:     msgcert,
		RepModCerts: modcertList,
		Mode:        mode,
	}
}
