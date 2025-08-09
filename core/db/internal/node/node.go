package node

import (
	"math/big"

	"github.com/devlup-labs/Libr/core/db/internal/models"
)

type DistanceNode struct {
	Node     *models.Node
	Distance *big.Int
}
