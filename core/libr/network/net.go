package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/libr-forum/Libr/core/libr/logger"
	Peers "github.com/libr-forum/Libr/core/libr/peers"
	"github.com/libr-forum/Libr/core/libr/types"
	util "github.com/libr-forum/Libr/core/libr/util"
)

type BaseResponse struct {
	Type string `json:"type"`
}

func SendTo(peerid string, route string, data interface{}, expect string) (interface{}, error) {

	switch expect {
	case "mod":
		switch v := data.(type) {
		case types.Msg:
			msgString, err := util.CanonicalizeMsg(v)
			if err != nil {
				log.Printf("Failed to generate canonical JSON: %v", err)
				return nil, err
			}
			resp, err := Peers.POST(peerid, route, []byte(msgString))
			if err != nil {
				return nil, err
			}
			var response types.ModCert
			if err := json.Unmarshal(resp, &response); err != nil {
				return nil, fmt.Errorf("bad mod response from %s: %w", peerid, err)
			}
			return response, nil
		case types.MsgCert:
			msgcertJSON, err := util.CanonicalizeMsgCert(v)
			if err != nil {
				log.Printf("Failed to generate canonical JSON: %v", err)
				return nil, err
			}
			resp, err := Peers.POST(peerid, route, []byte(msgcertJSON))
			if err != nil {
				return nil, err
			}
			var response types.ModCert
			if err := json.Unmarshal(resp, &response); err != nil {
				return nil, fmt.Errorf("bad mod response from %s: %w", peerid, err)
			}
			return response, nil
		default:
			logger.LogToFile("[DEBUG]Excpected msg or msgcert")
			return nil, errors.New("expected Msg or MsgCert struct for mod")
		}

	case "db":
		switch v := data.(type) {
		case types.MsgCert:
			msgcertJSON, err := util.CanonicalizeMsgCert(v)
			if err != nil {
				log.Printf("Failed to generate canonical JSON: %v", err)
				return nil, err
			}
			resp, err := Peers.POST(peerid, route, []byte(msgcertJSON))
			if err != nil {
				return nil, err
			}
			return resp, nil
		case types.ReportCert:
			reportCertJSON, err := util.CanonicalizeReportCert(v)
			if err != nil {
				log.Printf("Failed to generate canonical JSON: %v", err)
				return nil, err
			}
			resp, err := Peers.POST(peerid, route, []byte(reportCertJSON))
			if err != nil {
				return nil, err
			}
			return resp, nil
		default:
			return nil, errors.New("expected MsgCert or ReportCert struct for db")
		}

	default:
		return nil, errors.New("unknown response type requested")
	}
}

func GetFrom(peerid string, route string, key string) (interface{}, error) {

	resp, err := Peers.GET(peerid, route)
	if err != nil {
		return nil, err
	}
	return resp, nil

}
