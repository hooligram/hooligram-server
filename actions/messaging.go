package actions

import (
	"github.com/hooligram/hooligram-server/constants"
	"github.com/hooligram/hooligram-server/db"
)

///////////////////////
// MESSAGING_DELIVER //
///////////////////////

// MessagingDeliverRequest .
func MessagingDeliverRequest(actionID string, message *db.Message) *Action {
	payload := make(map[string]interface{})
	payload["content"] = message.Content
	payload["date_created"] = message.DateCreated
	payload["group_id"] = message.MessageGroupID
	payload["message_id"] = message.ID
	payload["sender_sid"] = message.SenderSID()

	return &Action{
		ID:      actionID,
		Payload: payload,
		Type:    constants.MessagingDeliverRequest,
	}
}

////////////////////
// MESSAGING_SEND //
////////////////////

// MessagingSendFailure .
func MessagingSendFailure(actionID string, errors []string) *Action {
	return constructFailureAction(actionID, constants.MessagingSendFailure, errors)
}

// MessagingSendSuccess .
func MessagingSendSuccess(actionID string, messageID int) *Action {
	payload := make(map[string]interface{})
	payload["action_id"] = actionID
	payload["message_id"] = messageID

	return &Action{
		ID:      actionID,
		Payload: payload,
		Type:    constants.MessagingSendSuccess,
	}
}
