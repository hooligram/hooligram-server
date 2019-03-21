package session

import (
	"github.com/gorilla/websocket"
	"github.com/hooligram/hooligram-server/structs"
	"github.com/hooligram/hooligram-server/utils"
)

var clients = make(map[*websocket.Conn]*structs.Client)

// Add .
func Add(conn *websocket.Conn) error {
	if _, ok := clients[conn]; ok {
		return nil
	}

	clients[conn] = &structs.Client{
		Conn:      conn,
		SessionID: utils.GenerateSessionID(),
	}

	return nil
}

// Get .
func Get(conn *websocket.Conn) (*structs.Client, bool) {
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

// SignIn .
func SignIn(
	conn *websocket.Conn,
	countryCode string,
	phoneNumber string,
	verificationCode string,
) *structs.Client {
	client, ok := Get(conn)
	if !ok {
		return nil
	}

	client.CountryCode = countryCode
	client.PhoneNumber = phoneNumber
	client.VerificationCode = verificationCode

	client.IsSignedIn = true

	return client
}
