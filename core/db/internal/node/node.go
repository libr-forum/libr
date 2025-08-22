package node

import (
	"math/big"

	"github.com/libr-forum/Libr/core/db/internal/models"
)

type DistanceNode struct {
	Node     *models.Node
	Distance *big.Int
}
