package main

import (
	"math/rand"
	"regexp"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/hooligram/logger"
)

func constructCreateGroupSuccessAction(
	groupID int64,
	groupName string,
	memberIDs []int,
	dateCreated string,
) *Action {
	payload := make(map[string]interface{})
	memberIDs = append([]int(nil), memberIDs...)

	payload["id"] = groupID
	payload["date_created"] = dateCreated
	payload["member_ids"] = memberIDs
	payload["name"] = groupName

	return &Action{
		Payload: payload,
		Type:    groupCreateSuccess,
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

func generateSessionID() string {
	var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	sessionID := make([]rune, 8)

	for i := range sessionID {
		sessionID[i] = runes[rand.Intn(len(runes))]
	}

	return string(sessionID)
}

func getDigits(s string) string {
	re := regexp.MustCompile("[^0-9]")
	return re.ReplaceAllString(s, "")
}

func logBody(filePath string, text string) {
	logger.Body(
		[]string{filePath},
		text,
	)
}

func logClose(client *Client, action *Action) {
	logger.Close(
		[]string{client.SessionID, strconv.Itoa(client.ID), action.Type},
		action.Payload,
	)
}

func logInfo(filePath string, text string) {
	logger.Info(
		[]string{filePath},
		text,
	)
}

func logOpen(client *Client, action *Action) {
	logger.Open(
		[]string{client.SessionID, strconv.Itoa(client.ID), action.Type},
		action.Payload,
	)
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
