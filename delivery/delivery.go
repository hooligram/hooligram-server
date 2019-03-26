package delivery

import (
	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/clients"
	"github.com/hooligram/hooligram-server/constants"
	"github.com/hooligram/hooligram-server/db"
	"github.com/hooligram/hooligram-server/utils"
)

// MessageDelivery .
type MessageDelivery struct {
	Message      *db.Message
	RecipientIDs []int
}

// MessageGroupDelivery .
type MessageGroupDelivery struct {
	MessageGroup *db.MessageGroup
	RecipientIDs []int
}

var (
	messageDeliveryChan      = make(chan *MessageDelivery)
	messageGroupDeliveryChan = make(chan *MessageGroupDelivery)
)

// DeliverMessage .
func DeliverMessage() {
	for {
		messageDelivery := <-GetMessageDeliveryChan()
		message := messageDelivery.Message
		recipientIDs := messageDelivery.RecipientIDs

		for _, client := range clients.GetSignedInClients() {
			if !utils.ContainsID(recipientIDs, client.GetID()) {
				continue
			}

			action := actions.MessagingDeliverRequest(message)
			action.ID = utils.GenerateRandomString(constants.ActionIDLength)
			client.WriteJSON(action)
		}
	}
}

// DeliverMessageGroup .
func DeliverMessageGroup() {
	for {
		messageGroupDelivery := <-GetMessageGroupDeliveryChan()
		messageGroup := messageGroupDelivery.MessageGroup
		recipientIDs := messageGroupDelivery.RecipientIDs

		for _, client := range clients.GetSignedInClients() {
			if !utils.ContainsID(recipientIDs, client.GetID()) {
				continue
			}

			action := actions.GroupDeliverRequest(messageGroup.ID)
			action.ID = utils.GenerateRandomString(constants.ActionIDLength)
			client.WriteJSON(action)
		}
	}
}

// GetMessageDeliveryChan .
func GetMessageDeliveryChan() chan *MessageDelivery {
	return messageDeliveryChan
}

// GetMessageGroupDeliveryChan .
func GetMessageGroupDeliveryChan() chan *MessageGroupDelivery {
	return messageGroupDeliveryChan
}
