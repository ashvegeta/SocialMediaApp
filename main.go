package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"socialmediaapp/src/routers"

	"github.com/gorilla/mux"
)

func main() {
	// Init Routers and Handlers
	r := mux.NewRouter()
	routers.UserRouter(r)

	// setup http server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	httpsrv := &http.Server{
		Handler: r,
		Addr:    ":" + port,
	}

	// start server
	fmt.Printf("listening at address: %s\n", httpsrv.Addr)
	err := httpsrv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
