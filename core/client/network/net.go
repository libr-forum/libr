package network

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/devlup-labs/Libr/core/client/types"
	util "github.com/devlup-labs/Libr/core/client/utils"
	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
)

func SendTo(addr string, data interface{}, expect string) (interface{}, error) {
	// Simulate network delay
	time.Sleep(300 * time.Millisecond)

	// Decode base64
	privBytes, _ := base64.StdEncoding.DecodeString("4TTsKvk1eUVdSibEDa5EKj30ecAZX7cVbTSNx/E9rkAwVAp84v7Zec7UitlhanRzE5Xs/gM3IQ5Ord1C+OBLmg==")
	priv := ed25519.PrivateKey(privBytes)

	switch expect {
	case "mod":
		msg, ok := data.(types.Msg)
		if !ok {
			return nil, errors.New("expected Msg struct for mod")
		}

		msgString, err := util.CanonicalizeMsg(msg)

		if err != nil {
			log.Printf("Failed to generate canonical JSON: %v", err)
			return nil, err
		}

		pubKeyStr, sign, err := cryptoutils.SignMessage(priv, msgString)
		if err != nil {
			return nil, err
		}

		status := "approved"
		if time.Now().Unix()%2 == 1 {
			status = "rejected"
		}
		fmt.Println("Status: ", status)
		response := types.ModCert{
			PublicKey: pubKeyStr,
			Sign:      sign,
			Status:    status,
		}
		return response, nil

	case "db":
		// simulate DB acknowledgement
		msgcert, ok := data.(types.MsgCert)
		if !ok {
			return nil, errors.New("expected MsgCert struct for db")
		}

		// Sort mod certs before verification
		sort.SliceStable(msgcert.ModCerts, func(i, j int) bool {
			return msgcert.ModCerts[i].PublicKey < msgcert.ModCerts[j].PublicKey
		})

		dataToVerify := types.DataToSign{
			Content:   msgcert.Msg.Content,
			Timestamp: msgcert.Msg.Ts,
			ModCerts:  msgcert.ModCerts,
		}

		jsonBytes, err := json.Marshal(dataToVerify)
		if err != nil {
			log.Printf("DB failed to marshal DataToSign: %v", err)
			return nil, err
		}

		if !cryptoutils.VerifySignature(msgcert.PublicKey, string(jsonBytes), msgcert.Sign) {
			return nil, fmt.Errorf("invalid MsgCert signature")
		}
		util.Store(msgcert)
		return "Message received and stored successfully", nil

	default:
		return nil, errors.New("unknown response type requested")
	}
}
