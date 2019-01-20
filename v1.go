package main

import (
	"log"
	"net/http"
)

func v1(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Can't upgrade to WebSocket connection.")
		return
	}

	defer conn.Close()

	for {
		action := Action{}
		err = conn.ReadJSON(&action)

		if err != nil {
			log.Println("Error reading JSON.")
			writeError(conn, 2001)
			continue
		}

		if action.Type == "" {
			log.Println("Action Type is required.")
			writeError(conn, 3001)
			continue
		}

		if action.Payload == nil {
			log.Println("Action Payload is required.")
			writeError(conn, 3001)
			continue
		}

		switch action.Type {
		case "VERIFICATION_VERIFY_PHONE_NO":
			conn.WriteJSON(Action{
				map[string]interface{}{},
				"VERIFICATION_VERIFICATION_CODE_SENT",
			})
		default:
			log.Println("Not supported Action Type.")
			writeError(conn, 3002)
		}
	}
}
