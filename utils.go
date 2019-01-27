package main

import "github.com/gorilla/websocket"

func writeError(conn *websocket.Conn, code int) {
	conn.WriteJSON(Action{
		map[string]interface{}{"code": code},
		"ERROR",
	})
}

func writeVerificationRequestCodeFailure(conn *websocket.Conn) {
	conn.WriteJSON(Action{
		map[string]interface{}{},
		verificationRequestCodeFailure,
	})
}
