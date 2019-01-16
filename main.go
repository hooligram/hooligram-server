package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type action struct {
	Payload interface{} `json:"payload"`
	Type    string      `json:"type"`
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/echo", echo)
	http.ListenAndServe(":8080", router)
}

func echo(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	action := action{}

	for {
		err = conn.ReadJSON(&action)
		if err != nil {
			continue
		}

		conn.WriteJSON(action)
	}
}
