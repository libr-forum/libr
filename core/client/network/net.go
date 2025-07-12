package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/devlup-labs/Libr/core/client/types"
	util "github.com/devlup-labs/Libr/core/client/util"
)

type BaseResponse struct {
	Type string `json:"type"`
}

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

		msgcertJSON, _ := util.CanonicalizeMsgCert(msgcert)

		resp, err := http.Post(addr, "application/json", bytes.NewBuffer([]byte(msgcertJSON)))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %v", err)
		}

		var base BaseResponse
		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&base); err != nil {
			return nil, fmt.Errorf("failed to decode base response: %v", err)
		}

		return bodyBytes, nil

	default:
		return nil, errors.New("unknown response type requested")
	}
}

func GetFrom(ip string, port string, route string, key string) (interface{}, error) {
	addr := fmt.Sprintf("http://%s:%s/%s?key=%s", ip, port, route, key)

	resp, err := http.Get(addr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %v", err)
	}

	var base BaseResponse
	if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&base); err != nil {
		return nil, fmt.Errorf("failed to decode base response: %v", err)
	}

	return bodyBytes, nil

}
