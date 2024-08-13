package routers

import (
	"socialmediaapp/src/handlers"

	"github.com/gorilla/mux"
)

func UserRouter(router *mux.Router) {
	router.HandleFunc("/user/add", handlers.AddUser).Methods("PUT")
}
