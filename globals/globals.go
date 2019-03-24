package globals

import (
	"github.com/hooligram/hooligram-server/db"
)

// MessageDelivery .
type MessageDelivery struct {
	Message      *db.Message
	RecipientIDs []int
}

var MessageDeliveryChan = make(chan *MessageDelivery)
