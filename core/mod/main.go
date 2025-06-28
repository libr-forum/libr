package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/devlup-labs/Libr/core/crypto/cryptoutils"
	"github.com/devlup-labs/Libr/core/mod/config"
	"github.com/devlup-labs/Libr/core/mod/internal/service"
	"github.com/devlup-labs/Libr/core/mod/models"
	"github.com/devlup-labs/Libr/core/mod/routers"
)

func main() {

	input := models.Msg{
		Content: "Hello, world!",
		Ts:      4234242,
	}
	fmt.Println(input)

	cryptoutils.GenerateKeyPair()

	load, nil := config.LoadConfig()
	fmt.Println(load, nil)

	r := routers.Routers()
	log.Fatal(http.ListenAndServe(":3000", r))

	clean, err := service.AnalyzeContent("helloo", service.AnalyzeWithKeywordFilter)
	if err != nil {
		fmt.Printf("OpenAI error: %v\n", err)
	} else {
		fmt.Printf("OpenAI result: %s\n", clean)
	}

}
