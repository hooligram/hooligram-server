package actions

import (
	"github.com/hooligram/hooligram-server/constants"
	"github.com/hooligram/hooligram-server/db"
)

//////////////////////////////
// CREATE_MESSAGING_DELIVER //
//////////////////////////////

// MessagingDeliverRequest .
func MessagingDeliverRequest(message *db.Message) *Action {
	payload := make(map[string]interface{})
	payload["content"] = message.Content
	payload["date_created"] = message.DateCreated
	payload["message_group_id"] = message.MessageGroupID
	payload["message_id"] = message.ID
	payload["sender_id"] = message.SenderID

	return &Action{
		Payload: payload,
		Type:    constants.MessagingDeliverRequest,
	}
}

// MessagingDeliverSuccessFailure .
func MessagingDeliverSuccessFailure(errors []string) *Action {
	return constructFailureAction(constants.MessagingDeliverSuccessFailure, errors)
}

// MessagingDeliverSuccessSuccess .
func MessagingDeliverSuccessSuccess() *Action {
	return constructEmptyAction(constants.MessagingDeliverSuccessSuccess)
}

///////////////////////////
// CREATE_MESSAGING_SEND //
///////////////////////////

// MessagingSendFailure .
func MessagingSendFailure(errors []string) *Action {
	return constructFailureAction(constants.MessagingSendFailure, errors)
}

// MessagingSendSuccess .
func MessagingSendSuccess(messageID int) *Action {
	payload := make(map[string]interface{})
	payload["message_id"] = messageID

	return &Action{
		Payload: payload,
		Type:    constants.MessagingSendSuccess,
	}
}
