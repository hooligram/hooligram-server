package main

import "net/http"

func addone(w http.ResponseWriter, r *http.Request) {
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
