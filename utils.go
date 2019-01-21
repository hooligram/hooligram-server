package main

import "github.com/gorilla/websocket"

func writeError(conn *websocket.Conn, code int) {
	conn.WriteJSON(Action{
		map[string]interface{}{"code": code},
		"ERROR",
	})
}
