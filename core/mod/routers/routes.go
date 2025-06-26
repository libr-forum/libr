package routers

import (
	"fmt"

	"github.com/devlup-labs/Libr/core/mod/internal/handlers"
	"github.com/gorilla/mux"
)

func Routers() *mux.Router {
	fmt.Println("setting up the routers")

	R := mux.NewRouter()
	R.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	R.HandleFunc("/submit", handlers.MsgIN).Methods("POST")            
	R.HandleFunc("/fetch/{timestamp}", handlers.MsgOUT).Methods("GET") 

	return R
}
