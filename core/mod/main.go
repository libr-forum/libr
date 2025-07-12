package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/devlup-labs/Libr/core/mod/config"
	"github.com/devlup-labs/Libr/core/mod/routers"
)

func main() {
	config.LoadConfig()

	r := routers.Routers()
	handlerWithCORS := routers.EnableCORS(r)
	fmt.Println("Listening on http://localhost:5000")
	log.Fatal(http.ListenAndServe(":3000", handlerWithCORS))
}
