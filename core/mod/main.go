package main

import (
	"fmt"

	"github.com/devlup-labs/Libr/core/mod/config"
	Peers "github.com/devlup-labs/Libr/core/mod/internal/peers"
)

func main() {
	load, _ := config.LoadConfig()
	fmt.Println(load)

	// r := routers.Routers()
	// fmt.Println("Listening on http://localhost:5000")
	// log.Fatal(http.ListenAndServe(":5000", r))

	relayAdd := ""
	Peers.StartNode(relayAdd)
}
