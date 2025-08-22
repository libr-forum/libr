package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"sort"

	"github.com/libr-forum/Libr/core/mod_client/keycache"
	"github.com/libr-forum/Libr/core/mod_client/logger"
	"github.com/libr-forum/Libr/core/mod_client/types"

	"github.com/libr-forum/Libr/core/crypto/cryptoutils"
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
	fmt.Println("data to sign = ", string(jsonBytes))
	fmt.Print("Private Key = ", base64.StdEncoding.EncodeToString(privKey))
	fmt.Println("Public Key = ", pubKeyStr)
	fmt.Print("Signature = ", sign)
	fmt.Print("\n")
	if err != nil {
		logger.LogToFile("[DEBUG]Failed to get sign message")
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
