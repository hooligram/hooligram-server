package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

// Action ...
type Action struct {
	Payload map[string]interface{} `json:"payload"`
	Type    string                 `json:"type"`
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

	defer conn.Close()
	action := Action{}

	for {
		err = conn.ReadJSON(&action)

		if err != nil {
			conn.WriteJSON(Action{
				map[string]interface{}{"code": 2001},
				"ERROR",
			})
			continue
		}

		conn.WriteJSON(action)
	}
}
