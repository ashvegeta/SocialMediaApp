package routers

import (
	"socialmediaapp/src/handlers"

	"github.com/gorilla/mux"
)

func UserRouter(router *mux.Router) {
	router.HandleFunc("/user/add", handlers.AddUser).Methods("PUT")
	router.HandleFunc("/user/delete", handlers.DelUser).Methods("POST")
	router.HandleFunc("/user/update", handlers.UpdateUserInfo).Methods("PUT")
}
