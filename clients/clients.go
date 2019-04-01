package clients

import (
	"github.com/gorilla/websocket"
	"github.com/hooligram/hooligram-server/constants"
	"github.com/hooligram/hooligram-server/utils"
)

const clientsTag = "clients"

var clients = make(map[*websocket.Conn]*Client)

// Add .
func Add(conn *websocket.Conn) error {
	if _, ok := clients[conn]; ok {
		return nil
	}

	clients[conn] = &Client{
		Conn:      conn,
		SessionID: generateSessionID(),
	}

	return nil
}

// Get .
func Get(conn *websocket.Conn) (*Client, bool) {
	client, ok := clients[conn]
	return client, ok
}

// GetSignedInClients .
func GetSignedInClients() []*Client {
	signedInclients := []*Client{}

	for _, client := range clients {
		if !client.isSignedIn {
			continue
		}

		signedInclients = append(signedInclients, client)
	}

	return signedInclients
}

// Remove .
func Remove(conn *websocket.Conn) error {
	if _, ok := clients[conn]; !ok {
		return nil
	}

	delete(clients, conn)
	return nil
}

func generateSessionID() string {
	return utils.GenerateRandomString(constants.SessionIDLength)
}
