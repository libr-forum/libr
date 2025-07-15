package network

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/devlup-labs/Libr/core/client/types"
	util "github.com/devlup-labs/Libr/core/client/util"
)

type BaseResponse struct {
	Type string `json:"type"`
}

func SendTo(ip string, port string, route string, data interface{}, expect string) (interface{}, error) {

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

		resp, err := Peers.POST(ip, port, route, []byte(msgcertJSON))
		if err != nil {
			return nil, err
		}

		return resp, nil

	default:
		return nil, errors.New("unknown response type requested")
	}
}

func GetFrom(ip string, port string, route string, key string) (interface{}, error) {

	resp, err := Peers.GET(ip, port, route)
	if err != nil {
		return nil, err
	}
	return resp, nil

}
