package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	peer "github.com/devlup-labs/Libr/core/db/internal/network/peers"
	"github.com/devlup-labs/Libr/core/db/internal/utils"
)

func main() {
	utils.SetupMongo("mongodb+srv://lemongrass8923:lmao1234@cluster0.vswojqe.mongodb.net/")
	relayAddrs, err := utils.GetRelayAddr()

	if err != nil {
		fmt.Println("Error while getting relay address, ", err)
	}
	fmt.Println(relayAddrs)

	peer.StartNode(relayAddrs)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("Interrupt received. Exiting gracefully.")
}
