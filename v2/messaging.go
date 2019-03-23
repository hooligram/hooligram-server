package v2

import (
	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/clients"
	"github.com/hooligram/hooligram-server/db"
	"github.com/hooligram/hooligram-server/globals"
	"github.com/hooligram/hooligram-server/utils"
)

func handleMessagingSendRequest(client *clients.Client, action *actions.Action) *actions.Action {
	content, ok := action.Payload["content"].(string)
	if !ok {
		failure := actions.CreateMessagingSendFailure([]string{"content not in payload"})
		client.WriteJSON(failure)
		return failure
	}

	messageGroupID, ok := action.Payload["message_group_id"].(float64)
	if !ok {
		failure := actions.CreateMessagingSendFailure([]string{"message_group_id not in payload"})
		client.WriteJSON(failure)
		return failure
	}

	senderID, ok := action.Payload["sender_id"].(float64)
	if !ok {
		failure := actions.CreateMessagingSendFailure([]string{"sender_id not in payload"})
		client.WriteJSON(failure)
		return failure
	}

	if client.GetID() != int(senderID) {
		failure := actions.CreateMessagingSendFailure([]string{"sender id mismatch"})
		client.WriteJSON(failure)
		return failure
	}

	if !db.IsClientInMessageGroup(int(senderID), int(messageGroupID)) {
		failure := actions.CreateMessagingSendFailure(
			[]string{"sender doesn't belong to message group"},
		)
		client.WriteJSON(failure)
		return failure
	}

	message, err := db.CreateMessage(content, int(messageGroupID), int(senderID))
	if err != nil {
		utils.LogBody(v2Tag, "error creating message. "+err.Error())
		failure := actions.CreateMessagingSendFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	messageGroupMemberIDs, err := db.FindAllMessageGroupMemberIDs(message.MessageGroupID)
	if err != nil {
		utils.LogBody(v2Tag, "error finding meesage group member ids. "+err.Error())
		failure := actions.CreateMessagingSendFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
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

	globals.MessageDeliveryChan <- &globals.MessageDelivery{
		Message:      message,
		RecipientIDs: recipientIDs,
	}

	success := actions.CreateMessagingSendSuccess(message.ID)
	client.WriteJSON(success)
	return success
}

func handleMessagingDeliverSuccess(client *clients.Client, action *actions.Action) {
	messageID, ok := action.Payload["message_id"].(float64)
	if !ok {
		client.WriteFailure(actions.MessagingDeliverFailure, []string{"message_id is missing"})
		return
	}

	recipientID := client.GetID()
	err := db.UpdateReceiptDateDelivered(int(messageID), recipientID)
	if err != nil {
		client.WriteFailure(actions.MessagingDeliverFailure, []string{err.Error()})
		return
	}
}
