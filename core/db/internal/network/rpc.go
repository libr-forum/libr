package network

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/models"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
	"github.com/devlup-labs/Libr/core/db/internal/storage"
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
	Ping(selfID [20]byte, selfPort string, target node.Node) error
}

type RealPinger struct{}

func (p *RealPinger) Ping(selfID [20]byte, selfPort string, target node.Node) error {
	return SendPing(selfID, selfPort, target)
}

func SendPing(selfID [20]byte, selfPort string, target node.Node) error {
	if GlobalPostFunc == nil {
		return fmt.Errorf("POST function not registered")
	}

	jsonStr := fmt.Sprintf(`{"node_id": "%x","port": "%s"}`, selfID[:], selfPort)
	jsonBytes := []byte(jsonStr)

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

func SendFindNode(targetId [20]byte, rt *routing.RoutingTable) []*node.Node {
	ClosestNodes := rt.FindClosest(targetId, config.K)
	return ClosestNodes
}

func StoreValue(key [20]byte, cert models.MsgCert, self *node.Node, rt *routing.RoutingTable) []*node.Node {
	closest := rt.FindClosest(key, config.K)
	fmt.Println(closest)

	selfDist := node.XORBigInt(self.NodeId, key)
	fmt.Println(selfDist)

	// Check if self is closer than any of the closest nodes
	for _, n := range closest {
		if selfDist.Cmp(node.XORBigInt(n.NodeId, key)) < 0 {
			storage.StoreMsgCert(cert)
			return nil
		}
	}

	return closest
}

func SendFindValue(key string, self *node.Node, rt *routing.RoutingTable) ([]models.MsgCert, []*node.Node) {
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

func DeleteValue(key [20]byte, cert models.MsgCert, self *node.Node, rt *routing.RoutingTable) []*node.Node {
	closest := rt.FindClosest(key, config.K)
	selfDist := node.XORBigInt(self.NodeId, key)

	for _, n := range closest {
		if selfDist.Cmp(node.XORBigInt(n.NodeId, key)) < 0 {
			storage.DeleteMsgCert(cert)
			return nil
		}
	}

	return closest
}
