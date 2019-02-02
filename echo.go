package main

import "net/http"

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
