package v2

import (
	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/clients"
	"github.com/hooligram/hooligram-server/db"
	"github.com/hooligram/hooligram-server/delivery"
	"github.com/hooligram/hooligram-server/utils"
)

/////////////////////////////////////
// HANDLER: MESSAGING_SEND_REQUEST //
/////////////////////////////////////

func handleMessagingSendRequest(client *clients.Client, action *actions.Action) *actions.Action {
	requestID := action.ID
	if requestID == "" {
		return messagingSendFailure(client, requestID, "id not in action")
	}

	content, ok := action.Payload["content"].(string)
	if !ok {
		return messagingSendFailure(client, requestID, "content not in payload")
	}

	messageGroupID, ok := action.Payload["message_group_id"].(float64)
	if !ok {
		return messagingSendFailure(client, requestID, "message_group_id not in payload")
	}

	senderID, ok := action.Payload["sender_id"].(float64)
	if !ok {
		return messagingSendFailure(client, requestID, "sender_id not in payload")
	}

	if client.GetID() != int(senderID) {
		return messagingSendFailure(client, requestID, "sender id mismatch")
	}

	if !db.ReadIsClientInMessageGroup(int(senderID), int(messageGroupID)) {
		return messagingSendFailure(client, requestID, "sender doesn't belong to message group")
	}

	message, err := db.CreateMessage(content, int(messageGroupID), int(senderID))
	if err != nil {
		utils.LogBody(v2Tag, "error creating message. "+err.Error())
		return messagingSendFailure(client, requestID, "server error")
	}

	messageGroupMemberIDs, err := db.ReadMessageGroupMemberIDs(message.MessageGroupID)
	if err != nil {
		utils.LogBody(v2Tag, "error finding meesage group member ids. "+err.Error())
		return messagingSendFailure(client, requestID, "server error")
	}

	var recipientIDs = make([]int, len(messageGroupMemberIDs))

	for i, recipientID := range messageGroupMemberIDs {
		if recipientID == int(message.SenderID) {
			continue
		}

		recipientIDs[i] = recipientID
	}

	for _, recipientID := range recipientIDs {
		db.CreateReceipt(message.ID, recipientID)
	}

	delivery.GetMessageDeliveryChan() <- &delivery.MessageDelivery{
		Message:      message,
		RecipientIDs: recipientIDs,
	}

	success := actions.MessagingSendSuccess(message.ID)
	success.ID = requestID
	client.WriteJSON(success)
	return success
}

////////////////////////////////////////
// HANDLER: MESSAGING_DELIVER_SUCCESS //
////////////////////////////////////////

func handleMessagingDeliverSuccess(client *clients.Client, action *actions.Action) *actions.Action {
	requestID := action.ID
	if requestID == "" {
		return messagingDeliverSuccessFailure(client, requestID, "id not in action")
	}

	messageID, ok := action.Payload["message_id"].(float64)
	if !ok {
		return messagingDeliverSuccessFailure(client, requestID, "message_id not in payload")
	}

	recipientID := client.GetID()
	ok, err := db.UpdateReceiptDateDelivered(int(messageID), recipientID)
	if err != nil {
		utils.LogBody(v2Tag, "error updating receipt date delivered. "+err.Error())
		return messagingDeliverSuccessFailure(client, requestID, "server error")
	}

	if !ok {
		return messagingDeliverSuccessFailure(client, requestID, "not allowed")
	}

	success := actions.MessagingDeliverSuccessSuccess(int(messageID))
	success.ID = requestID
	client.WriteJSON(success)
	return success
}

////////////
// HELPER //
////////////

func messagingSendFailure(client *clients.Client, actionID, err string) *actions.Action {
	failure := actions.MessagingSendFailure([]string{err})
	failure.ID = actionID
	client.WriteJSON(failure)
	return failure
}

func messagingDeliverSuccessFailure(client *clients.Client, actionID, err string) *actions.Action {
	failure := actions.MessagingDeliverSuccessFailure([]string{err})
	failure.ID = actionID
	client.WriteJSON(failure)
	return failure
}
