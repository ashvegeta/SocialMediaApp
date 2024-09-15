package routers

import (
	"socialmediaapp/src/handlers"

	"github.com/gorilla/mux"
)

func UserRouter(router *mux.Router) {
	router.HandleFunc("/user/add", handlers.AddUser).Methods("PUT")
	router.HandleFunc("/user/delete", handlers.DelUser).Methods("POST")
	router.HandleFunc("/user/update", handlers.UpdateUserInfo).Methods("PUT")
	router.HandleFunc("/conn/request", handlers.RequestConnection).Methods("PUT")
	router.HandleFunc("/conn/add", handlers.AddConnection).Methods("PUT")
	router.HandleFunc("/conn/delete", handlers.DelConnection).Methods("POST")
	router.HandleFunc("/notification/delete", handlers.DelNotification).Methods("POST")
	router.HandleFunc("/post/add", handlers.AddPost).Methods("PUT")
	router.HandleFunc("/search", handlers.Search).Methods("GET")
}
