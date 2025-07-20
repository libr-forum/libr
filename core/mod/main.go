package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
	"github.com/devlup-labs/Libr/core/mod/config"
	"github.com/devlup-labs/Libr/core/mod/routers"
)

func main() {

	cryptoutils.GenerateKeyPair()

	load, nil := config.LoadConfig()
	fmt.Println(load, nil)

	r := routers.Routers()
	log.Fatal(http.ListenAndServe(":3000", r))

}
