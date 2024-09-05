package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"socialmediaapp/src/routers"

	ghandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	// Init Routers and Handlers
	r := mux.NewRouter()
	routers.UserRouter(r)

	corsHandler := ghandlers.CORS(
		ghandlers.AllowedOrigins([]string{"*"}),                                       // Allow all origins; replace "*" with specific origins if needed
		ghandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}), // Allowed HTTP methods
		ghandlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),           // Allowed headers
	)

	// setup http server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	httpsrv := &http.Server{
		Handler: corsHandler(r),
		Addr:    ":" + port,
	}

	// start server
	fmt.Printf("listening at address: %s\n", httpsrv.Addr)
	err := httpsrv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
