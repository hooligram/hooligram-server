package actions

import "github.com/hooligram/hooligram-server/db"

//////////////////////////////
// CREATE_MESSAGING_DELIVER //
//////////////////////////////

// CreateMessagingDeliverRequest .
func CreateMessagingDeliverRequest(message *db.Message) *Action {
	payload := make(map[string]interface{})
	payload["content"] = message.Content
	payload["date_created"] = message.DateCreated
	payload["id"] = message.ID
	payload["sender_id"] = message.SenderID

	return &Action{
		Payload: payload,
		Type:    MessagingDeliverRequest,
	}
}

// CreateMessagingDeliverSuccessFailure .
func CreateMessagingDeliverSuccessFailure(errors []string) *Action {
	return createFailureAction(MessagingDeliverSuccessFailure, errors)
}

// CreateMessagingDeliverSuccessSuccess .
func CreateMessagingDeliverSuccessSuccess() *Action {
	return createEmptyAction(MessagingDeliverSuccessSuccess)
}

///////////////////////////
// CREATE_MESSAGING_SEND //
///////////////////////////

// CreateMessagingSendFailure .
func CreateMessagingSendFailure(errors []string) *Action {
	return createFailureAction(MessagingSendFailure, errors)
}

// CreateMessagingSendSuccess .
func CreateMessagingSendSuccess(messageID int) *Action {
	payload := make(map[string]interface{})
	payload["message_id"] = messageID

	return &Action{
		Payload: payload,
		Type:    MessagingSendSuccess,
	}
}
