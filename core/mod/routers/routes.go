package routers

import (
	"fmt"
	"net/http"

	"github.com/devlup-labs/Libr/core/mod/internal/handlers"
	"github.com/gorilla/mux"
)

func EnableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func Routers() *mux.Router {
	fmt.Println("setting up the routers")

	R := mux.NewRouter()
	R.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	R.HandleFunc("/submit", handlers.MsgIN).Methods("POST")
	//R.HandleFunc("/fetch/{timestamp}", handlers.MsgOUT).Methods("GET")

	return R
}
