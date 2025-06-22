package network

import (
	"errors"
	"fmt"
	"libr/keycache"
	"libr/types"
	util "libr/utils"
	"log"
	"time"

	"github.com/Arnav-Agrawal-987/crypto/cryptoutils"
)

func SendTo(addr string, data interface{}, expect string) (interface{}, error) {
	// Simulate network delay
	time.Sleep(300 * time.Millisecond)

	pub := keycache.PubKey
	priv := keycache.PrivKey

	switch expect {
	case "mod":
		msg, ok := data.(types.Msg)
		if !ok {
			return nil, errors.New("expected Msg struct for mod")
		}

		msgString, err := util.CanonicalizeMsg(msg)
		fmt.Println("🖊️  Signing this exact string:", msgString)

		if err != nil {
			log.Printf("Failed to generate canonical JSON: %v", err)
			return nil, err
		}

		fmt.Println("🔏 Mod is signing:", msgString)

		sign, err := cryptoutils.SignMessage(priv, msgString)
		if err != nil {
			return nil, err
		}

		status := "approved"
		if time.Now().Unix()%2 == 1 {
			status = "rejected"
		}

		response := types.ModCert{
			PublicKey: pub,
			Sign:      sign,
			Status:    status,
		}
		return response, nil

	case "db":
		// simulate DB acknowledgement
		return "Message received and stored successfully", nil

	default:
		return nil, errors.New("unknown response type requested")
	}
}
