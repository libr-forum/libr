package network

import (
	"encoding/json"
	"errors"
	"libr/types"
	"time"

	"github.com/Arnav-Agrawal-987/crypto/cryptoutils"
)

func SendTo(addr string, data interface{}, expect string) (interface{}, error) {
	// Simulate network delay
	time.Sleep(300 * time.Millisecond)

	pub, priv, err := cryptoutils.LoadKeys()
	if err != nil {
		return nil, err
	}

	// Serialize payload
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	payload := string(payloadBytes)

	// Sign the data
	sign, err := cryptoutils.SignMessage(priv, payload)
	if err != nil {
		return nil, err
	}

	// Response behavior based on type
	switch expect {
	case "mod":
		// simulate mod response
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
