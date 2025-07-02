package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/devlup-labs/Libr/core/mod/config"
	"github.com/devlup-labs/Libr/core/mod/internal/service"
	"github.com/devlup-labs/Libr/core/mod/routers"
)

func main() {
	load, _ := config.LoadConfig()
	fmt.Println(load)

	r := routers.Routers()
	fmt.Println("Listening on http://localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", r))

	clean, err := service.AnalyzeContent("helloo", service.AnalyzeWithKeywordFilter)
	if err != nil {
		fmt.Printf("OpenAI error: %v\n", err)
	} else {
		fmt.Printf("OpenAI result: %s\n", clean)
	}

}
