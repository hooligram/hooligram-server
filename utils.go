package main

import (
	"regexp"

	"github.com/gorilla/websocket"
)

func getDigits(s string) string {
	re := regexp.MustCompile("[^0-9]")
	return re.ReplaceAllString(s, "")
}

func constructBroadcastAction(source *Client, message string) *Action {
	sender := make(map[string]string)
	sender["country_code"] = source.CountryCode
	sender["phone_number"] = source.PhoneNumber

	payload := make(map[string]interface{})
	payload["message"] = message
	payload["sender"] = sender

	return &Action{
		Payload: payload,
		Type:    messagingBroadcastSuccess,
	}
}

func writeFailure(conn *websocket.Conn, actionType string, errors []string) {
	conn.WriteJSON(Action{
		map[string]interface{}{
			"errors": errors,
		},
		actionType,
	})
}

func writeEmptyAction(conn *websocket.Conn, actionType string) {
	conn.WriteJSON(Action{
		map[string]interface{}{},
		actionType,
	})
}
