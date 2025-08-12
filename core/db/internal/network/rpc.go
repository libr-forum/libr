package network

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/keycache"
	"github.com/devlup-labs/Libr/core/db/internal/models"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
	"github.com/devlup-labs/Libr/core/db/internal/storage"
	"github.com/devlup-labs/Libr/core/db/internal/utils"
)

var GlobalPinger Pinger
var GlobalPostFunc func(ip, port, route string, body []byte) ([]byte, error)

func RegisterPinger(p Pinger) {
	GlobalPinger = p
}

func RegisterPOST(f func(ip, port, route string, body []byte) ([]byte, error)) {
	GlobalPostFunc = f
}

type Pinger interface {
	Ping(selfID [20]byte, selfPort string, target *models.Node) error
}

type RealPinger struct{}

func (p *RealPinger) Ping(selfID [20]byte, selfPort string, target *models.Node) error {
	return SendPing(selfID, selfPort, target)
}

func SendPing(selfID [20]byte, selfPort string, target *models.Node) error {
	if GlobalPostFunc == nil {
		return fmt.Errorf("POST function not registered")
	}

	pubKeyB64 := base64.StdEncoding.EncodeToString(keycache.PubKey)
	jsonMap := map[string]string{
		"public_key": pubKeyB64,
	}
	jsonBytes, _ := json.Marshal(jsonMap)

	resp, err := GlobalPostFunc(target.IP, target.Port, "/route=ping", jsonBytes)
	if err != nil {
		fmt.Println("Ping failed:", err)
		return err
	}
	fmt.Println("Ping Response: ", string(resp))

	var res PingResponse
	if err := json.Unmarshal(resp, &res); err != nil {
		return err
	}

	if res.Status != "ok" {
		return errors.New("ping failed: not ok")
	}

	return nil
}

func SendFindNode(targetId [20]byte, rt *routing.RoutingTable) []*models.Node {
	ClosestNodes := rt.FindClosest(targetId, config.K)
	return ClosestNodes
}

func StoreValue(key [20]byte, cert *models.MsgCert, self *models.Node, rt *routing.RoutingTable) []*models.Node {
	closest := rt.FindClosest(key, config.K)
	fmt.Println(closest)

	selfDist := node.XORBigInt(self.NodeId, key)
	fmt.Println(selfDist)

	if len(closest) < config.K {
		storage.StoreMsgCert(cert)
		return nil
	}

	// Check if self is closer than any of the closest nodes
	for _, n := range closest {
		if selfDist.Cmp(node.XORBigInt(n.NodeId, key)) < 0 {
			storage.StoreMsgCert(cert)
			return nil
		}
	}

	return closest
}

func SendFindValue(key string, self *models.Node, rt *routing.RoutingTable) ([]models.RetMsgCert, []*models.Node) {
	ts, err := (strconv.ParseInt(key, 10, 64))
	if err != nil {
		return nil, nil
	}

	found := storage.GetMsgCert(ts)
	if len(found) > 0 {
		return found, nil
	}

	// Not found locally — return k closest to forward request
	keyBytes := sha1.Sum([]byte(key))
	return nil, rt.FindClosest(keyBytes, config.K)
}

// func DeleteValue(key *[20]byte, repCert *models.ReportCert, self *models.Node, rt *routing.RoutingTable) ([]*models.Node, error) {

// 	appr, rej := utils.CountModCerts(repCert.Msgcert.ModCerts)
// 	validMods, _ := utils.GetValidMods()
// 	modCount := len(repCert.Msgcert.ModCerts)
// 	shouldContinue := false
// 	var closest []*models.Node

// 	switch repCert.Mode {
// 	case "delete":
// 		if repCert.RepModCerts[0].Sign != repCert.Msgcert.Sign {
// 			return nil, fmt.Errorf("Wrong Mode")
// 		}
// 		closest = rt.FindClosest(*key, config.K)
// 		selfDist := node.XORBigInt(self.NodeId, *key)

// 		last := closest[len(closest)-1]
// 		lastDist := node.XORBigInt(last.NodeId, *key)

// 		if selfDist.Cmp(lastDist) < 0 {
// 			err := storage.DeleteMsgCert(repCert)
// 			if err != nil {
// 				if err.Error() != "MsgCert not found" {
// 					return nil, err
// 				}
// 			}
// 			return nil, nil
// 		}
// 	case "report":
// 		if err := storage.ValidateRepCert(repCert, validMods); err != nil {
// 			return nil, err
// 		}
// 		closest = rt.FindClosest(*key, config.K)
// 		selfDist := node.XORBigInt(self.NodeId, *key)

// 		last := closest[len(closest)-1]
// 		lastDist := node.XORBigInt(last.NodeId, *key)

// 		if selfDist.Cmp(lastDist) < 0 {
// 			err := storage.DeleteMsgCert(repCert)
// 			if err != nil {
// 				if err.Error() != "MsgCert not found" {
// 					return nil, err
// 				}
// 			}
// 			return nil, nil
// 		}
// 	}

// 	// switch {
// 	// case modCount == 0 && repCert.Msgcert.PublicKey == repCert.Msgcert.Msg.PublicKey:
// 	// 	shouldContinue = true

// 	// case modCount == 0 && repCert.PublicKey != repCert.ReportMsg.PublicKey:
// 	// 	return nil, fmt.Errorf("empty ModCert — invalid backup or post request")

// 	// case float32(appr)/float32(len(validMods)) <= config.MinApprove:
// 	// 	return nil, fmt.Errorf("less than 40% mod approval")

// 	// case appr < rej:
// 	// 	return nil, fmt.Errorf("more rejections than approvals")

// 	// case float32(appr)/float32(len(validMods)) > config.MinApprove && appr > rej:
// 	// 	shouldContinue = true
// 	// }

// 	// if !shouldContinue {
// 	// 	return nil, fmt.Errorf("less than 40% mod approval")
// 	// }

// 	return closest, nil
// }

func DeleteValue(key *[20]byte, repCert *models.ReportCert, self *models.Node, rt *routing.RoutingTable) ([]*models.Node, error) {
	validMods, _ := utils.GetOnlineMods()

	selfDist := node.XORBigInt(self.NodeId, *key)
	closest := rt.FindClosest(*key, config.K)

	close := false
	if len(closest) < config.K {
		close = true
	} else {
		lastDist := node.XORBigInt(closest[len(closest)-1].NodeId, *key)
		if selfDist.Cmp(lastDist) < 0 {
			close = true
		}
	}
	if close {
		// 🔒 Validate the full ReportCert (including MsgCert & RepModCerts)

		if err := storage.ValidateRepCert(repCert, validMods); err != nil {
			return nil, fmt.Errorf("repCert validation failed: %v", err)
		}
		err := storage.DeleteMsgCert(repCert)
		if err != nil && err.Error() != "MsgCert not found" {
			return nil, fmt.Errorf("deletion failed: %v", err)
		}
		return nil, nil
	}
	// 📡 Forward to closest peers
	return closest, nil
}
