package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/devlup-labs/Libr/core/mod/config"
	"github.com/devlup-labs/Libr/core/mod/internal/service"
	"github.com/devlup-labs/Libr/core/mod/models"
	"github.com/devlup-labs/Libr/core/mod/routers"
)

func main() {

	input := models.UserMsg{
		Content:   "Hello, world!",
		TimeStamp: 1749914634,
	}
	fmt.Println(input)

	//cryptoutils.LoadKeys()

	load, nil := config.LoadConfig()
	fmt.Println(load, nil)

	r := routers.Routers()
	log.Fatal(http.ListenAndServe(":5000", r))

	clean, err := service.AnalyzeContent("helloo", service.AnalyzeWithKeywordFilter)
	if err != nil {
		fmt.Printf("OpenAI error: %v\n", err)
	} else {
		fmt.Printf("OpenAI result: %s\n", clean)
	}

}
