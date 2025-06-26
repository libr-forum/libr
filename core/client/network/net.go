package network

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"

	"github.com/devlup-labs/Libr/core/client/types"
	util "github.com/devlup-labs/Libr/core/client/utils"
	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
)

// Map to simulate different mod behaviors
var modProbabilities = map[string]float64{
	"127.0.0.1:5000": 0.8, // Mod 0: mostly approves
	"127.0.0.1:5001": 0.5, // Mod 1: 50/50
	"127.0.0.1:5002": 0.2, // Mod 2: mostly rejects
}

// Simulated private key store
var ModPrivateKeys = map[string]string{
	"127.0.0.1:5000/mod": "uRG3nLqh2CHKMP2oRPndz2jeFa9rbGpVB4Eq6nY2LGFlcu+B1EoZPjrtj1AKPHJoS8bjRHK+Hic8OgeMthgToQ==",
	"127.0.0.1:5001/mod": "D+2jcJ42F5V/M71epF9NbnVFj9uIq+SAEKgjdXojI/S4tIskDH6egUB/PSZIjGidzpoPffq+ZKuA4PC2I3W2kg==",
	"127.0.0.1:5002/mod": "npUG5NkTCCd3x7HJa1A26OaFRCEWGGmCXl/tR1Jp+/++4Gd61sImlwcd0RiPxpkBgS+F/piDQ9lfCOz0Dlc2YA==",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func SendTo(ip string, port string, route string, data interface{}, expect string) (interface{}, error) {
	addr := fmt.Sprintf("%s:%s/%s", ip, port, route)

	// Simulate network delay
	time.Sleep(300 * time.Millisecond)

	switch expect {
	case "mod":
		privKeyStr, ok := ModPrivateKeys[addr]
		if !ok {
			return nil, fmt.Errorf("no private key found for %s", addr)
		}
		privBytes, err := base64.StdEncoding.DecodeString(privKeyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 private key for %s: %v", addr, err)
		}
		priv := ed25519.PrivateKey(privBytes)

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

		ipPort := fmt.Sprintf("%s:%s", ip, port)
		prob := modProbabilities[ipPort]
		status := "rejected"
		if rand.Float64() < prob {
			status = "approved"
		}

		fmt.Println("Status:", status)

		response := types.ModCert{
			PublicKey: pubKeyStr,
			Sign:      sign,
			Status:    status,
		}
		return response, nil

	case "db":
		msgcert, ok := data.(types.MsgCert)
		if !ok {
			return nil, errors.New("expected MsgCert struct for db")
		}

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
			return nil, fmt.Errorf("❌ Invalid MsgCert signature")
		}

		// ✅ From Version 1 — store the message
		util.Store(msgcert)

		return "Message received and stored successfully", nil

	default:
		return nil, errors.New("unknown response type requested")
	}
}
