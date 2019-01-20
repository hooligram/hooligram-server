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
	router.HandleFunc("/addone", addOne)
	router.HandleFunc("/echo", echo)
	http.ListenAndServe(":8080", router)
}

func addOne(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	defer conn.Close()

	for {
		action := Action{}
		err := conn.ReadJSON(&action)

		if err != nil {
			writeError(conn, 2001)
			continue
		}

		if action.Type != "ADD_ONE" {
			writeError(conn, 3002)
			continue
		}

		count, ok := action.Payload["count"].(float64)

		if !ok {
			writeError(conn, 3001)
			continue
		}

		count++
		action.Payload["count"] = count

		conn.WriteJSON(action)
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	defer conn.Close()

	for {
		action := Action{
			Payload: map[string]interface{}{},
			Type:    "",
		}
		err = conn.ReadJSON(&action)

		if err != nil {
			writeError(conn, 2001)
			continue
		}

		conn.WriteJSON(action)
	}
}

func writeError(conn *websocket.Conn, code int) {
	conn.WriteJSON(Action{
		map[string]interface{}{"code": code},
		"ERROR",
	})
}
