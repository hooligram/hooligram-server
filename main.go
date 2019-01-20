package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/addone", addone)
	router.HandleFunc("/echo", echo)
	http.ListenAndServe(":8080", router)
}
