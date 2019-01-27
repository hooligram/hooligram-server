package main

import "github.com/gorilla/websocket"

func writeEmptyAction(conn *websocket.Conn, actionType string) {
	conn.WriteJSON(Action{
		map[string]interface{}{},
		actionType,
	})
}

func writeError(conn *websocket.Conn, code int) {
	conn.WriteJSON(Action{
		map[string]interface{}{"code": code},
		"ERROR",
	})
}
