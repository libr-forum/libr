package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/devlup-labs/Libr/core/client/types"
	util "github.com/devlup-labs/Libr/core/client/utils"
	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
)

// Map to simulate different mod behaviors
// var modProbabilities = map[string]float64{
// 	"127.0.0.1:5000": 0.8, // Mod 0: mostly approves
// 	"127.0.0.1:5001": 0.5, // Mod 1: 50/50
// 	"127.0.0.1:5002": 0.2, // Mod 2: mostly rejects
// }

// Simulated private key store
// var ModPrivateKeys = map[string]string{
// 	"127.0.0.1:5000/mod": "uRG3nLqh2CHKMP2oRPndz2jeFa9rbGpVB4Eq6nY2LGFlcu+B1EoZPjrtj1AKPHJoS8bjRHK+Hic8OgeMthgToQ==",
// 	"127.0.0.1:5001/mod": "D+2jcJ42F5V/M71epF9NbnVFj9uIq+SAEKgjdXojI/S4tIskDH6egUB/PSZIjGidzpoPffq+ZKuA4PC2I3W2kg==",
// 	"127.0.0.1:5002/mod": "npUG5NkTCCd3x7HJa1A26OaFRCEWGGmCXl/tR1Jp+/++4Gd61sImlwcd0RiPxpkBgS+F/piDQ9lfCOz0Dlc2YA==",
// }

func SendTo(ip string, port string, route string, data interface{}, expect string) (interface{}, error) {
	addr := fmt.Sprintf("http://%s:%s/%s", ip, port, route)

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

		resp, err := http.Post(addr, "application/json", bytes.NewBuffer([]byte(msgString)))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var response types.ModCert

		json.NewDecoder(resp.Body).Decode(&response)

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
