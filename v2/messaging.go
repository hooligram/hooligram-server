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

	if !client.IsSignedIn() {
		return messagingSendFailure(client, requestID, "not signed in")
	}

	content, ok := action.Payload["content"].(string)
	if !ok {
		return messagingSendFailure(client, requestID, "content not in payload")
	}

	messageGroupID, ok := action.Payload["group_id"].(float64)
	if !ok {
		return messagingSendFailure(client, requestID, "group_id not in payload")
	}

	if !db.ReadIsClientInMessageGroup(int(client.GetID()), int(messageGroupID)) {
		return messagingSendFailure(client, requestID, "sender doesn't belong to message group")
	}

	message, err := db.CreateMessage(content, int(messageGroupID), client.GetID())
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

////////////
// HELPER //
////////////

func messagingSendFailure(client *clients.Client, actionID, err string) *actions.Action {
	failure := actions.MessagingSendFailure([]string{err})
	failure.ID = actionID
	client.WriteJSON(failure)
	return failure
}
