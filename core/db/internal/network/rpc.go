package network

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/devlup-labs/Libr/core/db/config"
	"github.com/devlup-labs/Libr/core/db/internal/models"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
	"github.com/devlup-labs/Libr/core/db/internal/storage"
)

type PingRequest struct {
	NodeID string `json:"node_id"`
	Port   string `json:"port"`
}

type PingResponse struct {
	Status string `json:"status"`
}

type RealPinger struct{}

func (p *RealPinger) Ping(selfID [20]byte, selfPort string, target node.Node) error {
	return SendPing(selfID, selfPort, target)
}

func SendPing(SelfID [20]byte, SelfPort string, node node.Node) error {
	addr := "http://" + node.IP + ":" + node.Port + "/ping"

	SelfIDhex := hex.EncodeToString(SelfID[:])
	body := map[string]string{
		"node_id": SelfIDhex,
		"port":    SelfPort,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := http.Post(addr, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var res map[string]string
	json.NewDecoder(resp.Body).Decode(&res)

	if res["status"] == "ok" {
		return nil
	}
	return errors.New("ping failed: status not ok")
}

func SendFindNode(targetId [20]byte, rt *routing.RoutingTable) []*node.Node {
	ClosestNodes := rt.FindClosest(targetId, config.K)
	return ClosestNodes
}

func StoreValue(key [20]byte, cert models.MsgCert, self *node.Node, rt *routing.RoutingTable) []*node.Node {
	closest := rt.FindClosest(key, config.K)

	for _, n := range closest {
		if n.IP == self.IP && n.Port == self.Port {
			// Only store if current node is one of the k closest
			storage.StoreMsgCert(cert)
			return nil
		}
	}
	return closest
}

func SendFindValue(key string, self *node.Node, rt *routing.RoutingTable) ([]models.MsgCert, []*node.Node) {
	ts, err := strconv.ParseInt(key, 10, 64)
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
