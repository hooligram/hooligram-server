package delivery

import "github.com/hooligram/hooligram-server/db"

// MessageDelivery .
type MessageDelivery struct {
	Message      *db.Message
	RecipientIDs []int
}

var messageDeliveryChan = make(chan *MessageDelivery)

// GetMessageDeliveryChan .
func GetMessageDeliveryChan() chan *MessageDelivery {
	return messageDeliveryChan
}
