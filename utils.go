package main

import (
	"regexp"

	"github.com/gorilla/websocket"
)

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

func constructDeliverMessageAction(message *Message) *Action {
	payload := make(map[string]interface{})
	payload["content"] = message.Content
	payload["date_created"] = message.DateCreated
	payload["id"] = message.ID
	payload["sender_id"] = message.SenderID

	return &Action{
		Payload: payload,
		Type:    messagingDeliverRequest,
	}
}

func containsID(ids []int, id int) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}

	return false
}

func getDigits(s string) string {
	re := regexp.MustCompile("[^0-9]")
	return re.ReplaceAllString(s, "")
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
