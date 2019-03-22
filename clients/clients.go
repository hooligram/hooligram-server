package clients

import (
	"math/rand"

	"github.com/gorilla/websocket"
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

// Remove .
func Remove(conn *websocket.Conn) error {
	if _, ok := clients[conn]; !ok {
		return nil
	}

	delete(clients, conn)
	return nil
}

func generateSessionID() string {
	var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	sessionID := make([]rune, 8)

	for i := range sessionID {
		sessionID[i] = runes[rand.Intn(len(runes))]
	}

	return string(sessionID)
}
