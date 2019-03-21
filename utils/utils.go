package utils

import (
	"math/rand"
	"regexp"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/hooligram/hooligram-server/constants"
	"github.com/hooligram/hooligram-server/structs"
	"github.com/hooligram/logger"
)

// ConstructCreateGroupSuccessAction .
func ConstructCreateGroupSuccessAction(
	groupID int64,
	groupName string,
	memberIDs []int,
	dateCreated string,
) *structs.Action {
	payload := make(map[string]interface{})
	memberIDs = append([]int(nil), memberIDs...)

	payload["id"] = groupID
	payload["date_created"] = dateCreated
	payload["member_ids"] = memberIDs
	payload["name"] = groupName

	return &structs.Action{
		Payload: payload,
		Type:    constants.GroupCreateSuccess,
	}
}

// ConstructDeliverMessageAction .
func ConstructDeliverMessageAction(message *structs.Message) *structs.Action {
	payload := make(map[string]interface{})
	payload["content"] = message.Content
	payload["date_created"] = message.DateCreated
	payload["id"] = message.ID
	payload["sender_id"] = message.SenderID

	return &structs.Action{
		Payload: payload,
		Type:    constants.MessagingDeliverRequest,
	}
}

// ContainsID .
func ContainsID(ids []int, id int) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}

	return false
}

// GenerateSessionID .
func GenerateSessionID() string {
	var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	sessionID := make([]rune, 8)

	for i := range sessionID {
		sessionID[i] = runes[rand.Intn(len(runes))]
	}

	return string(sessionID)
}

// GetDigits .
func GetDigits(s string) string {
	re := regexp.MustCompile("[^0-9]")
	return re.ReplaceAllString(s, "")
}

// LogBody .
func LogBody(filePath string, text string) {
	logger.Body(
		[]string{filePath},
		text,
	)
}

// LogClose .
func LogClose(client *structs.Client, action *structs.Action) {
	logger.Close(
		[]string{client.SessionID, strconv.Itoa(client.ID), action.Type},
		action.Payload,
	)
}

// LogInfo .
func LogInfo(filePath string, text string) {
	logger.Info(
		[]string{filePath},
		text,
	)
}

// LogOpen .
func LogOpen(client *structs.Client, action *structs.Action) {
	logger.Open(
		[]string{client.SessionID, strconv.Itoa(client.ID), action.Type},
		action.Payload,
	)
}

// WriteFailure .
func WriteFailure(conn *websocket.Conn, actionType string, errors []string) {
	conn.WriteJSON(structs.Action{
		map[string]interface{}{
			"errors": errors,
		},
		actionType,
	})
}

// WriteEmptyAction .
func WriteEmptyAction(conn *websocket.Conn, actionType string) {
	conn.WriteJSON(structs.Action{
		map[string]interface{}{},
		actionType,
	})
}
