package client

import (
	"crypto/ed25519"
	"encoding/json"
	"log"
	"sort"

	"github.com/libr-forum/Libr/core/libr/keycache"
	"github.com/libr-forum/Libr/core/libr/logger"
	"github.com/libr-forum/Libr/core/libr/types"

	"github.com/libr-forum/Libr/core/libr/crypto/cryptoutils"
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
		ModCerts:  modcertList, // sorted before signing
	}

	jsonBytes, _ := json.Marshal(dataToSign)
	pubKeyStr, sign, err := cryptoutils.SignMessage(privKey, string(jsonBytes))
	if err != nil {
		// Not fatal: an unsigned cert is rejected by every DB node.
		logger.LogToFile("[DEBUG]Failed to get sign message")
		log.Printf("failed to sign message: %v", err)
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

func CreateRepCert(msgcert types.MsgCert, modcertList []types.ModCert, mode string) types.ReportCert {
	sort.SliceStable(modcertList, func(i, j int) bool {
		return modcertList[i].PublicKey < modcertList[j].PublicKey
	})

	repCert := types.ReportCert{
		Msgcert:     msgcert,
		RepModCerts: modcertList,
		Mode:        mode,
	}

	return repCert
}
