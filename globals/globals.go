package globals

import (
	"github.com/gorilla/websocket"
	"github.com/hooligram/hooligram-server/db"
)

// MessageDelivery .
type MessageDelivery struct {
	Message      *db.Message
	RecipientIDs []int
}

var MessageDeliveryChan = make(chan *MessageDelivery)
var TwilioAPIKey string
var Upgrader = websocket.Upgrader{}
