package server

import (
	"net/http"

	"github.com/devlup-labs/Libr/core/db/internal/handlers"
	"github.com/devlup-labs/Libr/core/db/internal/node"
	"github.com/devlup-labs/Libr/core/db/internal/routing"
)

func SetupRoutes(localNode *node.Node, rt *routing.RoutingTable) {
	http.Handle("/ping", withCORS(http.HandlerFunc(handlers.HandlePing(localNode, rt))))
	http.Handle("/find_node", withCORS(http.HandlerFunc(handlers.FindNodeHandler(localNode, rt))))
	http.Handle("/store", withCORS(http.HandlerFunc(handlers.StoreHandler(localNode, rt))))
	http.Handle("/find_value", withCORS(http.HandlerFunc(handlers.FindValueHandler(localNode, rt))))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
