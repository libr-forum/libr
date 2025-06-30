package server

import (
	"net/http"

	"github.com/devlup-labs/Libr/core/db/internal/handlers"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
)

func SetupRoutes(localNode *node.Node, rt *routing.RoutingTable) {
	http.HandleFunc("/ping", handlers.HandlePing(localNode, rt))
	http.HandleFunc("/find_node", handlers.FindNodeHandler(localNode, rt))
	http.HandleFunc("/store", handlers.StoreHandler(localNode, rt))
	http.HandleFunc("/find_value", handlers.FindValueHandler(localNode, rt))
}
