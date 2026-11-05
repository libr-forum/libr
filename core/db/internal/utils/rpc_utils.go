package utils

import (
	"github.com/libr-forum/Libr/core/db/config"
	"github.com/libr-forum/Libr/core/db/internal/models"
	"github.com/libr-forum/Libr/core/db/internal/node"
	"github.com/libr-forum/Libr/core/db/internal/routing"
)

func CountModCerts(modCerts []models.ModCert) (approved, rejected int) {
	for _, cert := range modCerts {
		switch cert.Status {
		case "1":
			approved++
		case "0":
			rejected++
		}
	}
	return approved, rejected
}

func ShouldDelete(self *models.Node, key *[20]byte, rt *routing.RoutingTable) bool {
	closest := rt.FindClosest(*key, config.K)
	if len(closest) == 0 {
		return false
	}
	selfDist := node.XORBigInt(self.NodeId, *key)
	lastDist := node.XORBigInt(closest[len(closest)-1].NodeId, *key)
	return selfDist.Cmp(lastDist) < 0
}
